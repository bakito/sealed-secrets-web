package main

import (
	_ "embed"

	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ricoberger/sealed-secrets-web/pkg/handler"
	"github.com/ricoberger/sealed-secrets-web/pkg/marshal"
	"github.com/ricoberger/sealed-secrets-web/pkg/secrets"
)

var (
	//go:embed static/index.html
	indexTemplate string
	// go:embed static
	static embed.FS
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}
func main() {

	kubesealArgs := "--cert=testdata/cert.pem"
	outputFormat := "yaml"
	disableLoadSecrets := true
	webExternalUrl := ""

	m := marshal.For(outputFormat)
	sealer := secrets.Seal(kubesealArgs)

	indexHTML := renderIndexHTML(outputFormat, disableLoadSecrets, webExternalUrl)

	h := handler.New(indexHTML, m, sealer, disableLoadSecrets)

	r := gin.Default()
	r.GET("/", h.Index)
	r.StaticFS("/static", http.FS(static))
	r.GET("/_health", h.Health)

	api := r.Group("/api")
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	api.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	api.POST("/seal", h.Seal)
	api.POST("/encode", h.Encode)
	api.POST("/decode", h.Decode)

	r.NoRoute(h.RedirectToIndex)
	_ = r.Run()
}

func renderIndexHTML(outputFormat string, disableLoadSecrets bool, webExternalUrl string) string {
	indexTmpl := template.Must(template.New("index.html").Parse(indexTemplate))
	data := map[string]interface{}{
		"OutputFormat":       outputFormat,
		"DisableLoadSecrets": disableLoadSecrets,
		"WebExternalUrl":     webExternalUrl,
	}
	var tpl bytes.Buffer
	indexTmpl.Execute(&tpl, data)
	indexHTML := tpl.String()
	return indexHTML
}
