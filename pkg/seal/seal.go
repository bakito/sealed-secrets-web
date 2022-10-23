package seal

import (
	"bytes"
	"context"
	"crypto/rsa"
	"log"
	"os"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	"github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/bitnami-labs/sealed-secrets/pkg/kubeseal"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

type Sealer interface {
	Secret(secret string) ([]byte, error)
	Raw(data Raw) ([]byte, error)
}

var _ Sealer = &apiSealer{}

func NewAPISealer(ss config.SealedSecrets) (Sealer, error) {
	log.Printf("Connection to sealed secrets with (%s)\n", ss.String())

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	cc := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, nil, os.Stdout)

	f, err := kubeseal.OpenCert(context.TODO(), cc, ss.Namespace, ss.Service, ss.CertURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	pubKey, err := kubeseal.ParseKey(f)
	if err != nil {
		return nil, err
	}

	return &apiSealer{
		clientConfig: cc,
		pubKey:       pubKey,
	}, nil
}

type apiSealer struct {
	clientConfig clientcmd.ClientConfig
	pubKey       *rsa.PublicKey
}

func (a *apiSealer) Secret(secret string) ([]byte, error) {
	var buf bytes.Buffer
	if err := kubeseal.Seal(
		a.clientConfig,
		"json",
		strings.NewReader(secret),
		&buf,
		scheme.Codecs,
		a.pubKey,
		v1alpha1.DefaultScope,
		false,
		"",
		"",
	); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (a *apiSealer) Raw(data Raw) ([]byte, error) {
	var buf bytes.Buffer
	scope := v1alpha1.DefaultScope
	if data.Scope != "" {
		_ = scope.Set(data.Scope)
	}
	if err := kubeseal.EncryptSecretItem(
		&buf, data.Name, data.Namespace, []byte(data.Value),
		v1alpha1.DefaultScope, a.pubKey); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type Raw struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Scope     string `json:"scope"`
}
