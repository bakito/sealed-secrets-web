package seal

import (
	"bytes"
	"context"
	"crypto/rsa"
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

const DefaultCertRefreshInterval = 5 * time.Minute

type Sealer interface {
	Raw(ctx context.Context, data Raw) ([]byte, error)
	Certificate(ctx context.Context) ([]byte, error)
	Seal(ctx context.Context, outputFormat string, secret io.Reader) ([]byte, error)
	Validate(ctx context.Context, secret io.Reader) error
}

var _ Sealer = &apiSealer{}

func NewAPISealer(ctx context.Context, ss config.SealedSecrets, refreshInterval time.Duration) (Sealer, error) {
	log.Printf("Connection to sealed secrets with (%s)\n", ss.String())

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	cc := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, nil, os.Stdout)

	if refreshInterval <= 0 {
		refreshInterval = DefaultCertRefreshInterval
	}

	pubKey, err := fetchPublicKey(ctx, cc, ss)
	if err != nil {
		return nil, err
	}

	return &apiSealer{
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

type apiSealer struct {
	mu              sync.Mutex
	clientConfig    clientcmd.ClientConfig
	ss              config.SealedSecrets
	pubKey          *rsa.PublicKey
	lastRefresh     time.Time
	refreshInterval time.Duration
}

func (a *apiSealer) getPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if time.Since(a.lastRefresh) >= a.refreshInterval {
		pubKey, err := fetchPublicKey(ctx, a.clientConfig, a.ss)
		if err != nil {
			log.Printf("Failed to refresh public key, falling back to cached key: %v", err)
			return a.pubKey, nil
		}
		a.pubKey = pubKey
		a.lastRefresh = time.Now()
		log.Print("Public key refreshed successfully")
	}

	return a.pubKey, nil
}

func (a *apiSealer) Certificate(ctx context.Context) ([]byte, error) {
	f, err := kubeseal.OpenCert(ctx, a.clientConfig, a.ss.Namespace, a.ss.Service, a.ss.CertURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a *apiSealer) Seal(ctx context.Context, outputFormat string, secret io.Reader) ([]byte, error) {
	pubKey, err := a.getPublicKey(ctx)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = kubeseal.Seal(
		a.clientConfig,
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

func (a *apiSealer) Raw(ctx context.Context, data Raw) ([]byte, error) {
	pubKey, err := a.getPublicKey(ctx)
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

func (a *apiSealer) Validate(ctx context.Context, secret io.Reader) error {
	return kubeseal.ValidateSealedSecret(
		ctx,
		a.clientConfig,
		a.ss.Namespace,
		a.ss.Service,
		secret,
	)
}

type Raw struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Scope     string `json:"scope"`
}
