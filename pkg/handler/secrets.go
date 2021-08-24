package handler

/*
func secretsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if *disableLoadSecrets {
			http.Error(w, fmt.Sprintf("Loading secrets is disabled"), http.StatusForbidden)
			return
		}

		// List all secrets.
		sec, err := sHandler.List()

		js, err := json.Marshal(sec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
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
			_, _ = w.Write(js)
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
		_, _ = w.Write(js)
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
		_, _ = w.Write(js)
		return
	}

	http.Error(w, "invalid method", http.StatusInternalServerError)
}
*/
