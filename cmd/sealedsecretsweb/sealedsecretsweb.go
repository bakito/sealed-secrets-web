package main

import (
	goflag "flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/ricoberger/sealed-secrets-web/pkg/secrets"
	"github.com/ricoberger/sealed-secrets-web/pkg/version"

	"github.com/bitnami-labs/flagenv"
	"github.com/bitnami-labs/pflagenv"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	flagEnvPrefix = "SEALED_SECRETS"
)

var (
	certFile       = flag.String("cert", "", "Certificate / public key to use for encryption. Overrides --controller-*")
	controllerNs   = flag.String("controller-namespace", metav1.NamespaceSystem, "Namespace of sealed-secrets controller.")
	controllerName = flag.String("controller-name", "sealed-secrets-controller", "Name of sealed-secrets controller.")
	disableLoadSecrets = flag.Bool("disable-load-secrets", false, "Disable the loading of existing secrets")
	outputFormat   = flag.String("format", "json", "Output format for sealed secret. Either json or yaml")
	printVersion   = flag.Bool("version", false, "Print version information and exit")

	clientConfig clientcmd.ClientConfig
	sHandler     *secrets.Handler

	indexTmpl = template.Must(template.ParseFiles("static/index.html"))
)

func init() {
	flagenv.SetFlagsFromEnv(flagEnvPrefix, goflag.CommandLine)

	// The "usual" clientcmd/kubectl flags
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	flag.StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, flag.CommandLine, kflags)
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)

	pflagenv.SetFlagsFromEnv(flagEnvPrefix, flag.CommandLine)

	// Standard goflags (glog in particular)
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

func main() {
	// We are using the same flags as the kubeseal command-line tool.
	// The flags are passed to the kubeseal client to seal the secrets.
	flag.Parse()
	goflag.CommandLine.Parse([]string{})

	if *printVersion {
		v, err := version.Print("sealed secrets web")
		if err != nil {
			log.Fatalf("Could not get version information: %s", err.Error())
		}

		fmt.Println(v)
		return
	}

	var err error
	sHandler, err = secrets.NewHandler(clientConfig, *outputFormat)
	if err != nil {
		log.Fatalf("Could not initialize secrets handler: %s", err.Error())
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/_health", healthHandler)
	http.HandleFunc("/api/seal", sealHandler)
	http.HandleFunc("/api/secrets", secretsHandler)
	http.HandleFunc("/api/base64", base64Handler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Start the web interface.
	fmt.Printf("Starting \"Sealed Secrets\" Web Interface %s\n", version.Info())
	fmt.Printf("Build context %s\n", version.BuildContext())

	fmt.Printf("Listening on %s\n", ":8080")
	log.Println(http.ListenAndServe(":8080", nil))
}
