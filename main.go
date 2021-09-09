package main

import (
	"bytes"
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/handler"
	"github.com/bakito/sealed-secrets-web/pkg/marshal"
	"github.com/bakito/sealed-secrets-web/pkg/seal"
	"github.com/bakito/sealed-secrets-web/pkg/secrets"
	"github.com/bakito/sealed-secrets-web/pkg/version"
	"github.com/gin-gonic/gin"
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

	disableLoadSecrets = flag.Bool("disable-load-secrets", false, "Disable the loading of existing secrets")
	kubesealArgs       = flag.String("kubeseal-arguments", "", "Arguments which are passed to kubeseal")
	outputFormat       = flag.String("format", "json", "Output format for sealed secret. Either json or yaml")
	initialSecretFile  = flag.String("initial-secret-file", "", "Define a file with the initial secret to be displayed. If empty, defaults are used.")
	webExternalUrl     = flag.String("web-external-url", "", "The URL under which the Sealed Secrets Web Interface is externally reachable (for example, if it is served via a reverse proxy).")
	printVersion       = flag.Bool("version", false, "Print version information and exit")
	port               = flag.Int("port", 8080, "Define the port to run the application on. (default: 8080)")
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

	flag.Parse()

	if *printVersion {
		fmt.Println(version.Print("sealed secrets web"))
		return
	}

	log.Printf("Running sealed secrets web (%s) on port %d", version.Version, *port)
	_ = setupRouter().Run(fmt.Sprintf(":%d", *port))
}

func setupRouter() *gin.Engine {
	m := marshal.For(*outputFormat)
	sealer := seal.New(*kubesealArgs)

	indexHTML, err := renderIndexHTML(*outputFormat, *disableLoadSecrets, *webExternalUrl)
	if err != nil {
		log.Fatalf("Could not render the index html template: %s", err.Error())
	}
	sHandler, err := secrets.NewHandler(clientConfig, *outputFormat, *disableLoadSecrets)
	if err != nil {
		log.Fatalf("Could not initialize secrets handler: %s", err.Error())
	}

	r := gin.New()
	r.Use(gin.Recovery())
	h := handler.New(indexHTML, m, sealer)

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

	r.NoRoute(h.RedirectToIndex)
	return r
}

func renderIndexHTML(outputFormat string, disableLoadSecrets bool, webExternalUrl string) (string, error) {
	indexTmpl := template.Must(template.New("index.html").Parse(indexTemplate))
	initialSecret := initialSecretJSON
	if initialSecretFile != nil && strings.TrimSpace(*initialSecretFile) != "" {
		b, err := ioutil.ReadFile(*initialSecretFile)
		if err != nil {
			return "", err
		}
		initialSecret = string(b)
	} else if strings.EqualFold(outputFormat, "yaml") {
		initialSecret = initialSecretYAML
	}

	data := map[string]interface{}{
		"OutputFormat":       outputFormat,
		"DisableLoadSecrets": disableLoadSecrets,
		"WebExternalUrl":     webExternalUrl,
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
