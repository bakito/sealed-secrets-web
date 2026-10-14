package seal

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bitnami/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/bitnami/sealed-secrets/pkg/kubeseal"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/bakito/sealed-secrets-web/pkg/config"
)

const (
	DefaultCertRefreshInterval = 5 * time.Minute
	defaultKeyLabel            = "default"
)

type Sealer interface {
	Raw(ctx context.Context, data Raw) ([]byte, error)
	Certificate(ctx context.Context) ([]byte, error)
	Seal(ctx context.Context, outputFormat string, secret io.Reader) ([]byte, error)
	Validate(ctx context.Context, secret io.Reader) error
}

type SealerWithKeys interface {
	Sealer
	Keys() []string
	SealWithKey(ctx context.Context, keyLabel, outputFormat string, secret io.Reader) ([]byte, error)
	RawWithKey(ctx context.Context, keyLabel string, data Raw) ([]byte, error)
	CertificateWithKey(ctx context.Context, keyLabel string) ([]byte, error)
}

var _ SealerWithKeys = &apiSealer{}

func NewAPISealer(
	ctx context.Context,
	ss config.SealedSecrets,
	keys []config.SealedSecretsKey,
	refreshInterval time.Duration,
) (SealerWithKeys, error) {
	log.Printf("Connection to sealed secrets with (%s)\n", ss.String())

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	cc := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, nil, os.Stdout)

	if refreshInterval <= 0 {
		refreshInterval = DefaultCertRefreshInterval
	}

	// Create the default sealer
	defaultSealer, err := newSingleSealer(ctx, cc, ss, refreshInterval)
	if err != nil {
		return nil, err
	}

	sealers := map[string]*singleSealer{
		defaultKeyLabel: defaultSealer,
	}

	// Create additional sealers for configured keys
	for _, key := range keys {
		if key.Label == "" {
			return nil, errors.New("sealedSecretsKeys entry missing label")
		}
		ssKey := config.SealedSecrets{
			Service:   key.Service,
			Namespace: key.Namespace,
			CertURL:   key.CertURL,
		}
		s, err := newSingleSealer(ctx, cc, ssKey, refreshInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to create sealer for key %q: %w", key.Label, err)
		}
		sealers[key.Label] = s
		log.Printf("Registered sealing key: %s (%s)", key.Label, ssKey.String())
	}

	return &apiSealer{
		sealers: sealers,
	}, nil
}

func newSingleSealer(
	ctx context.Context,
	cc clientcmd.ClientConfig,
	ss config.SealedSecrets,
	refreshInterval time.Duration,
) (*singleSealer, error) {
	pubKey, err := fetchPublicKey(ctx, cc, ss)
	if err != nil {
		return nil, err
	}

	return &singleSealer{
		clientConfig:    cc,
		ss:              ss,
		pubKey:          pubKey,
		lastRefresh:     time.Now(),
		refreshInterval: refreshInterval,
	}, nil
}

func fetchPublicKey(ctx context.Context, cc clientcmd.ClientConfig, ss config.SealedSecrets) (*rsa.PublicKey, error) {
	f, err := kubeseal.OpenCert(ctx, cc, ss.Namespace, ss.Service, ss.CertURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return kubeseal.ParseKey(f)
}

type singleSealer struct {
	mu              sync.Mutex
	clientConfig    clientcmd.ClientConfig
	ss              config.SealedSecrets
	pubKey          *rsa.PublicKey
	lastRefresh     time.Time
	refreshInterval time.Duration
}

func (s *singleSealer) getPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Since(s.lastRefresh) >= s.refreshInterval {
		pubKey, err := fetchPublicKey(ctx, s.clientConfig, s.ss)
		if err != nil {
			log.Printf("Failed to refresh public key, falling back to cached key: %v", err)
			return s.pubKey, nil
		}
		s.pubKey = pubKey
		s.lastRefresh = time.Now()
		log.Print("Public key refreshed successfully")
	}

	return s.pubKey, nil
}

func (s *singleSealer) certificate(ctx context.Context) ([]byte, error) {
	f, err := kubeseal.OpenCert(ctx, s.clientConfig, s.ss.Namespace, s.ss.Service, s.ss.CertURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(f)
}

func (s *singleSealer) seal(ctx context.Context, outputFormat string, secret io.Reader) ([]byte, error) {
	pubKey, err := s.getPublicKey(ctx)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = kubeseal.Seal(
		s.clientConfig,
		outputFormat,
		secret,
		&buf,
		scheme.Codecs,
		pubKey,
		v1alpha1.DefaultScope,
		false,
		"",
		"",
	)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *singleSealer) raw(ctx context.Context, data Raw) ([]byte, error) {
	pubKey, err := s.getPublicKey(ctx)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	scope := v1alpha1.DefaultScope
	if data.Scope != "" {
		_ = scope.Set(data.Scope)
	}
	err = kubeseal.EncryptSecretItem(
		&buf, data.Name, data.Namespace, []byte(data.Value),
		scope, pubKey)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *singleSealer) validate(ctx context.Context, secret io.Reader) error {
	return kubeseal.ValidateSealedSecret(
		ctx,
		s.clientConfig,
		s.ss.Namespace,
		s.ss.Service,
		secret,
	)
}

type apiSealer struct {
	sealers map[string]*singleSealer
}

func (a *apiSealer) getSealer(keyLabel string) *singleSealer {
	if keyLabel == "" {
		keyLabel = defaultKeyLabel
	}
	s, ok := a.sealers[keyLabel]
	if !ok {
		return a.sealers[defaultKeyLabel]
	}
	return s
}

func (a *apiSealer) Keys() []string {
	keys := make([]string, 0, len(a.sealers))
	for k := range a.sealers {
		keys = append(keys, k)
	}
	return keys
}

func (a *apiSealer) SealWithKey(ctx context.Context, keyLabel, outputFormat string, secret io.Reader) ([]byte, error) {
	return a.getSealer(keyLabel).seal(ctx, outputFormat, secret)
}

func (a *apiSealer) RawWithKey(ctx context.Context, keyLabel string, data Raw) ([]byte, error) {
	return a.getSealer(keyLabel).raw(ctx, data)
}

func (a *apiSealer) CertificateWithKey(ctx context.Context, keyLabel string) ([]byte, error) {
	return a.getSealer(keyLabel).certificate(ctx)
}

// Seal uses the default key.
func (a *apiSealer) Seal(ctx context.Context, outputFormat string, secret io.Reader) ([]byte, error) {
	return a.SealWithKey(ctx, defaultKeyLabel, outputFormat, secret)
}

// Raw uses the default key.
func (a *apiSealer) Raw(ctx context.Context, data Raw) ([]byte, error) {
	return a.RawWithKey(ctx, defaultKeyLabel, data)
}

// Certificate returns the certificate of the default key.
func (a *apiSealer) Certificate(ctx context.Context) ([]byte, error) {
	return a.CertificateWithKey(ctx, defaultKeyLabel)
}

// Validate uses the default key.
func (a *apiSealer) Validate(ctx context.Context, secret io.Reader) error {
	return a.getSealer(defaultKeyLabel).validate(ctx, secret)
}

type Raw struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Scope     string `json:"scope"`
}
