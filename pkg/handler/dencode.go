package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/bitnami-labs/sealed-secrets/pkg/multidocyaml"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func (h *Handler) Dencode(c *gin.Context) {
	outputContentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	secret, err := readSecret(scheme.Codecs.UniversalDecoder(), c.Request.Body)
	if err != nil {
		log.Printf("Error in %s: %s\n", Sanitize(c.Request.URL.Path), Sanitize(err.Error()))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	encode, err := encodeSecret(h.dencode(secret), outputFormat)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, outputContentType, encode)
}

func (h *Handler) dencode(secret *v1.Secret) *v1.Secret {
	if secret.StringData != nil && len(secret.StringData) > 0 {
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		for key, value := range secret.StringData {
			secret.Data[key] = []byte(value)
		}
		secret.StringData = nil
		return secret
	}

	if secret.Data != nil && len(secret.Data) > 0 {
		if secret.StringData == nil {
			secret.StringData = map[string]string{}
		}
		for key, value := range secret.Data {
			secret.StringData[key] = string(value)
		}
		secret.Data = nil
	}
	return secret
}

func readSecret(codec runtime.Decoder, r io.Reader) (*v1.Secret, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := multidocyaml.EnsureNotMultiDoc(data); err != nil {
		return nil, err
	}

	var ret v1.Secret
	if err = runtime.DecodeInto(codec, data, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}
