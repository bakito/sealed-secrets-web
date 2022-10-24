package handler

import (
	"path"
	"strings"
)

func Sanitize(urlPath string) string {
	p := path.Clean(urlPath)
	p = strings.ReplaceAll(p, "\n", "")
	p = strings.ReplaceAll(p, "\r", "")
	return p
}
