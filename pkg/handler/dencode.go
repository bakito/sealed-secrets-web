package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/bitnami/sealed-secrets/pkg/multidocyaml"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func (h *Handler) Dencode(c *gin.Context) {
	outputContentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error in %s: %s\n", Sanitize(c.FullPath()), Sanitize(err.Error()))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := validateBase64Data(body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	secret, err := readSecret(scheme.Codecs.UniversalDecoder(), bytes.NewReader(body))
	if err != nil {
		log.Printf("Error in %s: %s\n", Sanitize(c.FullPath()), Sanitize(err.Error()))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	encode, err := encodeSecret(h.dencodeInternal(secret), outputFormat)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.FullPath()), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, outputContentType, encode)
}

func (*Handler) dencodeInternal(secret *corev1.Secret) *corev1.Secret {
	if len(secret.StringData) > 0 {
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		for key, value := range secret.StringData {
			secret.Data[key] = []byte(value)
		}
		secret.StringData = nil
		return secret
	}

	if len(secret.Data) > 0 {
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

func readSecret(codec runtime.Decoder, r io.Reader) (*corev1.Secret, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := multidocyaml.EnsureNotMultiDoc(data); err != nil {
		return nil, err
	}

	var ret corev1.Secret
	if err := runtime.DecodeInto(codec, data, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}
