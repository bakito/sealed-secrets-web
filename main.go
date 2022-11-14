package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	"github.com/bakito/sealed-secrets-web/pkg/handler"
	"github.com/bakito/sealed-secrets-web/pkg/seal"
	"github.com/bakito/sealed-secrets-web/pkg/secrets"
	"github.com/bakito/sealed-secrets-web/pkg/version"
	ssClient "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealedsecrets/v1alpha1"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	//go:embed templates/index.html
	indexTemplate string
	//go:embed templates/secret.json
	initialSecretJSON string
	//go:embed templates/secret.yaml
	initialSecretYAML string

	//go:embed static/*
	static      embed.FS
	staticFS, _ = fs.Sub(static, "static")

	clientConfig clientcmd.ClientConfig
)

func init() {
	gin.SetMode(gin.ReleaseMode)

	// The "usual" clientcmd/kubectl flags
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("Could not read the config: %s", err.Error())
	}

	if cfg.PrintVersion {
		fmt.Println(version.Print("sealed secrets web"))
		return
	}

	coreClient, ssc, err := secrets.BuildClients(clientConfig, cfg.DisableLoadSecrets)
	if err != nil {
		log.Fatalf("Could build k8s clients:%v", err.Error())
	}
	sealer, err := seal.NewAPISealer(cfg.SealedSecrets)
	if err != nil {
		log.Fatalf("Setup sealer: %s", err.Error())
	}

	log.Printf("Running sealed secrets web (%s / %s) on port %d", version.Version, cfg.OutputFormat, cfg.Web.Port)
	_ = setupRouter(coreClient, ssc, cfg, sealer).Run(fmt.Sprintf(":%d", cfg.Web.Port))
}

func setupRouter(coreClient corev1.CoreV1Interface, ssClient ssClient.BitnamiV1alpha1Interface, cfg *config.Config, sealer seal.Sealer) *gin.Engine {
	indexHTML, err := renderIndexHTML(cfg)
	if err != nil {
		log.Fatalf("Could not render the index html template: %s", err.Error())
	}

	sHandler := secrets.NewHandler(coreClient, ssClient, cfg)

	r := gin.New()
	r.Use(gin.Recovery())
	if cfg.Web.Logger {
		r.Use(gin.LoggerWithFormatter(ginLogFormatter()))
	}
	h := handler.New(indexHTML, sealer, cfg)

	r.GET("/", h.Index)
	r.StaticFS("/static", http.FS(staticFS))
	r.GET("/_health", h.Health)

	api := r.Group("/api")

	api.GET("/version", h.Version)
	api.POST("/seal", h.Seal)
	api.POST("/raw", h.Raw)
	api.GET("/certificate", h.Certificate)
	api.POST("/kubeseal", h.KubeSeal)
	api.POST("/dencode", h.Dencode)
	api.POST("/encode", h.Encode)
	api.POST("/decode", h.Decode)

	api.GET("/secret/:namespace/:name", sHandler.Secret)
	api.GET("/secrets", sHandler.AllSecrets)

	r.NoRoute(h.RedirectToIndex(cfg.Web.Context))
	return r
}

func renderIndexHTML(cfg *config.Config) (string, error) {
	indexTmpl := template.Must(template.New("index.html").Parse(indexTemplate))
	initialSecret := initialSecretJSON
	if cfg.InitialSecret != "" {
		initialSecret = cfg.InitialSecret
	} else if strings.EqualFold(cfg.OutputFormat, "yaml") {
		initialSecret = initialSecretYAML
	}

	data := map[string]interface{}{
		"OutputFormat":       cfg.OutputFormat,
		"DisableLoadSecrets": cfg.DisableLoadSecrets,
		"WebContext":         cfg.Web.Context,
		"InitialSecret":      initialSecret,
		"Version":            version.Version,
	}

	var tpl bytes.Buffer
	if err := indexTmpl.Execute(&tpl, data); err != nil {
		return "", err
	}
	indexHTML := tpl.String()
	return indexHTML, nil
}

func ginLogFormatter() func(param gin.LogFormatterParams) string {
	return func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			handler.Sanitize(param.Path),
			param.ErrorMessage,
		)
	}
}
