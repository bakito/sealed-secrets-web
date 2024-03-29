package handler

const (
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
