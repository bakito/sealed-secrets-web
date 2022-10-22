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

	log.Printf("Running sealed secrets web (%s) on port %d", version.Version, cfg.Web.Port)
	_ = setupRouter(coreClient, ssc, cfg).Run(fmt.Sprintf(":%d", cfg.Web.Port))
}

func setupRouter(coreClient corev1.CoreV1Interface, ssClient ssClient.BitnamiV1alpha1Interface, cfg *config.Config) *gin.Engine {
	sealer := seal.New(cfg.KubesealArgs)

	indexHTML, err := renderIndexHTML(cfg)
	if err != nil {
		log.Fatalf("Could not render the index html template: %s", err.Error())
	}

	sHandler := secrets.NewHandler(coreClient, ssClient, cfg)

	r := gin.New()
	r.Use(gin.Recovery())
	h := handler.New(indexHTML, sealer, cfg)

	r.GET("/", h.Index)
	r.StaticFS("/static", http.FS(staticFS))
	r.GET("/_health", h.Health)

	api := r.Group("/api")

	api.GET("/version", h.Version)
	api.POST("/seal", h.Seal)
	api.POST("/raw", h.Raw)
	api.POST("/encode", h.Encode)
	api.POST("/decode", h.Decode)

	api.GET("/secret/:namespace/:name", sHandler.Secret)
	api.GET("/secrets", sHandler.AllSecrets)

	r.NoRoute(h.RedirectToIndex(cfg.Web.ExternalURL))
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
		"WebExternalUrl":     cfg.Web.ExternalURL,
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
