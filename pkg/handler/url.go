package handler

import (
	"path"
	"strings"
)

func Sanitize(value string) string {
	p := path.Clean(value)
	p = strings.ReplaceAll(p, "\n", "")
	p = strings.ReplaceAll(p, "\r", "")
	return p
}
