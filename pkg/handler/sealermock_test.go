package handler

import (
	"fmt"
	"io"

	"github.com/bakito/sealed-secrets-web/pkg/seal"
)

var (
	validCertificate = `-----BEGIN CERTIFICATE-----
[...]
-----END CERTIFICATE-----
`
	encryptedRawValue = "encryptedRawValue"
	sealAsJSON        = `{"apiVersion": "bitnami.com/v1alpha1"}`
	sealedAsYAML      = "apiVersion: bitnami.com/v1alpha1\n"

	stringDataAsJSON = `{
  "kind": "Secret",
  "apiVersion": "v1",
  "metadata": {
    "creationTimestamp": null
  },
  "stringData": {
    "username": "admin"
  },
  "type": "Opaque"
}`
	dataAsJSON = `{
  "kind": "Secret",
  "apiVersion": "v1",
  "metadata": {
    "creationTimestamp": null
  },
  "data": {
    "username": "YWRtaW4="
  },
  "type": "Opaque"
}`
	stringDataAsYAML = `apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
stringData:
  username: admin
type: Opaque
`
	dataAsYAML = `apiVersion: v1
data:
  username: YWRtaW4=
kind: Secret
metadata:
  creationTimestamp: null
type: Opaque
`
)

type successfulSealer struct{}

func (m successfulSealer) Raw(_ seal.Raw) ([]byte, error) {
	return []byte(encryptedRawValue), nil
}

func (m successfulSealer) Seal(outputFormat string, _ io.Reader) ([]byte, error) {
	if outputFormat == "json" {
		return []byte(sealAsJSON), nil
	}
	if outputFormat == "yaml" {
		return []byte(sealedAsYAML), nil
	}
	return nil, fmt.Errorf("unknown format")
}

func (m successfulSealer) Certificate() ([]byte, error) {
	return []byte(validCertificate), nil
}

type errorSealer struct{}

func (m errorSealer) Raw(_ seal.Raw) ([]byte, error) {
	return nil, fmt.Errorf("unexpected error")
}

func (m errorSealer) Seal(_ string, _ io.Reader) ([]byte, error) {
	return nil, fmt.Errorf("unexpected error")
}

func (m errorSealer) Certificate() ([]byte, error) {
	return nil, fmt.Errorf("unexpected error")
}
