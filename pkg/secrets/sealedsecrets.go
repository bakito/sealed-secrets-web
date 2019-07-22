package secrets

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	ssClient "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealed-secrets/v1alpha1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/cert"
)

// SealedSecretsHandler handles the sealing of secrets.
// The function for the SealedSecretsHandler are taken from the 'kubeseal' client.
// The 'kubeseal' client can be found in the repository for sealed secrets: https://github.com/bitnami-labs/sealed-secrets/blob/master/cmd/kubeseal/main.go
type SealedSecretsHandler struct {
	clientConfig   clientcmd.ClientConfig
	certFile       string
	controllerNs   string
	controllerName string
	outputFormat   string
	pubKey         *rsa.PublicKey
}

// NewSealedSecretsHandler creates a new handler for sealing secrets.
func NewSealedSecretsHandler(clientConfig clientcmd.ClientConfig, certFile string, controllerNs string, controllerName string, outputFormat string) (*SealedSecretsHandler, error) {
	handler := &SealedSecretsHandler{
		clientConfig:   clientConfig,
		certFile:       certFile,
		controllerNs:   controllerNs,
		controllerName: controllerName,
		outputFormat:   outputFormat,
	}

	// Read the certificate and parse the public key.
	f, err := handler.openCert()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pubKey, err := handler.parseKey(f)
	if err != nil {
		return nil, err
	}

	handler.pubKey = pubKey

	return handler, nil
}

func (h *SealedSecretsHandler) parseKey(r io.Reader) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	certs, err := cert.ParseCertsPEM(data)
	if err != nil {
		return nil, err
	}

	// ParseCertsPem returns error if len(certs) == 0, but best to be sure...
	if len(certs) == 0 {
		return nil, errors.New("Failed to read any certificates")
	}

	cert, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Expected RSA public key but found %v", certs[0].PublicKey)
	}

	return cert, nil
}

func (h *SealedSecretsHandler) readSecret(codec runtime.Decoder, data []byte) (*v1.Secret, error) {
	var ret v1.Secret
	if err := runtime.DecodeInto(codec, data, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (h *SealedSecretsHandler) prettyEncoder(codecs runtimeserializer.CodecFactory, mediaType string, gv runtime.GroupVersioner) (runtime.Encoder, error) {
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, fmt.Errorf("binary can't serialize %s", mediaType)
	}

	prettyEncoder := info.PrettySerializer
	if prettyEncoder == nil {
		prettyEncoder = info.Serializer
	}

	enc := codecs.EncoderForVersion(prettyEncoder, gv)
	return enc, nil
}

func (h *SealedSecretsHandler) openCertFile(certFile string) (io.ReadCloser, error) {
	f, err := os.Open(certFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %v", certFile, err)
	}

	return f, nil
}

func (h *SealedSecretsHandler) openCertHTTP(c corev1.CoreV1Interface, namespace, name string) (io.ReadCloser, error) {
	f, err := c.Services(namespace).ProxyGet("http", name, "", "/v1/cert.pem", nil).Stream()
	if err != nil {
		return nil, fmt.Errorf("Error fetching certificate: %v", err)
	}

	return f, nil
}

func (h *SealedSecretsHandler) openCert() (io.ReadCloser, error) {
	if h.certFile != "" {
		return h.openCertFile(h.certFile)
	}

	conf, err := h.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	conf.AcceptContentTypes = "application/x-pem-file, */*"
	restClient, err := corev1.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return h.openCertHTTP(restClient, h.controllerNs, h.controllerName)
}

func (h *SealedSecretsHandler) sealedSecretOutput(codecs runtimeserializer.CodecFactory, ssecret *ssv1alpha1.SealedSecret) ([]byte, error) {
	var contentType string
	switch strings.ToLower(h.outputFormat) {
	case "json", "":
		contentType = runtime.ContentTypeJSON
	case "yaml":
		contentType = "application/yaml"
	default:
		return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
	}
	prettyEnc, err := h.prettyEncoder(codecs, contentType, ssv1alpha1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}
	buf, err := runtime.Encode(prettyEnc, ssecret)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Seal handles the sealing of a Kubernetes secret.
// This function creates the sealed secret.
func (h *SealedSecretsHandler) Seal(data []byte, codecs runtimeserializer.CodecFactory) ([]byte, error) {
	// Parse the secret.
	secret, err := h.readSecret(codecs.UniversalDecoder(), data)
	if err != nil {
		return nil, err
	}

	if len(secret.Data) == 0 {
		return nil, fmt.Errorf("Secret.data is empty in input Secret, assuming this is an error and aborting")
	}

	if secret.GetName() == "" {
		return nil, fmt.Errorf("Missing metadata.name in input Secret")
	}

	if secret.GetNamespace() == "" {
		ns, _, err := h.clientConfig.Namespace()
		if err != nil {
			return nil, err
		}
		secret.SetNamespace(ns)
	}

	// Strip read-only server-side ObjectMeta (if present)
	secret.SetSelfLink("")
	secret.SetUID("")
	secret.SetResourceVersion("")
	secret.Generation = 0
	secret.SetCreationTimestamp(metav1.Time{})
	secret.SetDeletionTimestamp(nil)
	secret.DeletionGracePeriodSeconds = nil

	ssecret, err := ssv1alpha1.NewSealedSecret(codecs, h.pubKey, secret)
	if err != nil {
		return nil, err
	}

	return h.sealedSecretOutput(codecs, ssecret)
}

// List returns a list of all secrets.
func (h *SealedSecretsHandler) List() ([]Secret, error) {
	conf, err := h.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := ssClient.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	ssList, err := client.SealedSecrets("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var secrets []Secret

	for _, item := range ssList.Items {
		secret := Secret{}
		secret.MetaData.Name = item.Name
		secret.MetaData.Namespace = item.Namespace

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// GetSecret returns a secret by name in the given namespace.
func (h *SealedSecretsHandler) GetSecret(namespace, name string) ([]byte, error) {
	conf, err := h.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	restClient, err := corev1.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	secret, err := restClient.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var data map[string]string
	data = make(map[string]string)

	for key, value := range secret.Data {
		data[key] = string(value)
	}

	s := Secret{}
	s.Kind = secret.TypeMeta.Kind
	s.APIVersion = secret.TypeMeta.APIVersion
	s.MetaData.Name = secret.Name
	s.MetaData.Namespace = secret.Namespace
	s.MetaData.Labels = secret.Labels
	s.MetaData.Annotations = secret.Annotations
	s.Data = data
	s.StringData = secret.StringData
	s.Type = secret.Type

	if h.outputFormat == "yaml" {
		return yaml.Marshal(s)
	} else if h.outputFormat == "json" {
		return json.Marshal(s)
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}
