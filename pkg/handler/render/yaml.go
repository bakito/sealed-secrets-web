// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"net/http"

	"gopkg.in/yaml.v3"
)

// create custom render because gin still uses yaml v2 => remove when gin starts using "gopkg.in/yaml.v3"

// YAML contains the given interface object.
type YAML struct {
	Data any
}

var yamlContentType = []string{"application/x-yaml; charset=utf-8"}

// Render (YAML) marshals the given interface object and writes data with custom ContentType.
func (r YAML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	yamlEncoder := yaml.NewEncoder(w)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(r.Data)
	if err != nil {
		return err
	}
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (r YAML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, yamlContentType)
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
