package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ricoberger/sealed-secrets-web/pkg/secrets"
	"gopkg.in/yaml.v2"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		OutputFormat       string
		DisableLoadSecrets bool
		WebExternalUrl     string
	}{
		*outputFormat,
		*disableLoadSecrets,
		*webExternalUrl,
	}

	indexTmpl.Execute(w, data)
}

func sealHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Secret string `json:"secret"`
	}{
		"",
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ss, err := secrets.Seal(data.Secret, *kubesealArgs)
	if err != nil {
		http.Error(w, fmt.Sprintf("kubeseal error: %s\n\n%s", err.Error(), string(ss)), http.StatusBadRequest)
		return
	}

	if *outputFormat == "yaml" {
		// unmarshal result to json
		sec := make(map[string]interface{})
		if err := json.Unmarshal(ss, &sec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// marshal to yaml
		if ss, err = yaml.Marshal(sec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	data.Secret = string(ss)

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func secretsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if *disableLoadSecrets {
			http.Error(w, fmt.Sprintf("Loading secrets is disabled"), http.StatusForbidden)
			return
		}

		// List all secrets.
		secrets, err := sHandler.List()

		js, err := json.Marshal(secrets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	} else if r.Method == http.MethodPost {
		// Encode / Decode the 'data' field of a secret.
		data := struct {
			Encode bool   `json:"encode"`
			Secret string `json:"secret"`
		}{
			false,
			"",
		}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if data.Encode == true {
			encoded, err := sHandler.Encode(data.Secret)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			data.Secret = string(encoded)

			js, err := json.Marshal(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
			return
		}

		decoded, err := sHandler.Decode(data.Secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data.Secret = string(decoded)

		js, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	} else if r.Method == http.MethodPut {
		if *disableLoadSecrets {
			http.Error(w, fmt.Sprintf("Loading secrets is disabled"), http.StatusForbidden)
			return
		}

		// Load existing secret.
		data := struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Secret    string `json:"secret"`
		}{
			"",
			"",
			"",
		}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		secret, err := sHandler.GetSecret(data.Namespace, data.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data.Secret = string(secret)

		js, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	http.Error(w, "invalid method", http.StatusInternalServerError)
}

func base64Handler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Secret string `json:"secret"`
		Encode bool   `json:"encode"`
	}{
		"",
		false,
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Encode/Decode secret data

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
