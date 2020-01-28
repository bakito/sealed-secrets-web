package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ricoberger/sealed-secrets-web/pkg/secrets"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	return
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		OutputFormat       string
		DisableLoadSecrets bool
	}{
		*outputFormat,
		*disableLoadSecrets,
	}

	indexTmpl.Execute(w, data)
	return
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

	ss, err := secrets.Seal(data.Secret)
	if err != nil {
		http.Error(w, fmt.Sprintf("kubeseal error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	data.Secret = string(ss)

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return
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
	return
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
	return
}
