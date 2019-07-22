package main

import (
	goflag "flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/ricoberger/sealed-secrets-web/pkg/secrets"
	"github.com/ricoberger/sealed-secrets-web/pkg/version"

	"github.com/bitnami-labs/flagenv"
	"github.com/bitnami-labs/pflagenv"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	flagEnvPrefix = "SEALED_SECRETS"

	yamlExample = `# The 'apiVersion' and 'kind' should always be 'v1' and 'Secret'.
apiVersion: v1
kind: Secret
# Metadata section of the secret.
# The encoding and decoding function only uses the 'name', 'namespace', 'annotations'
# and 'labels' field.
# All other fields will be striped during the encoding / decoding.
metadata:
  name: mysecretname
  namespace: mysecretnamespace
# All fields in the 'data' section will be encoded, decoded or encryped.
data:
  username: admin
  password: admin
  values.yaml: |
    secretName: mysecretname
    secretValue: mysecretvalue
    subSecrets:
      key: value
# The type of the Secret can be any valid Kubernetes secret type.
# Normaly this should be 'Opaque'.
type: Opaque`

	jsonExample = `{
  "apiVersion": "v1",
  "kind": "Secret",
  "metadata": {
    "name": "mysecretname",
    "namespace": "mysecretnamespace"
  },
  "data": {
    "username": "admin",
    "password": "admin",
    "values.yaml": "secretName: mysecretname\nsecretValue: mysecretvalue\nsubSecrets:\n  key: value\n"
  },
  "type": "Opaque"
}`
)

var (
	listenAddress  = flag.String("listen-address", ":8080", "Address to listen on for web interface.")
	showVersion    = flag.Bool("version", false, "Show version information.")
	certFile       = flag.String("cert", "", "Certificate / public key to use for encryption. Overrides --controller-*")
	controllerNs   = flag.String("controller-namespace", metav1.NamespaceSystem, "Namespace of sealed-secrets controller.")
	controllerName = flag.String("controller-name", "sealed-secrets-controller", "Name of sealed-secrets controller.")
	outputFormat   = flag.String("format", "yaml", "Output format for sealed secret. Either json or yaml")

	clientConfig clientcmd.ClientConfig
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
	// Parse command-line flags.
	flag.Parse()
	goflag.CommandLine.Parse([]string{})

	// Show version information.
	if *showVersion {
		v, err := version.Print("sealed-secrets-web")
		if err != nil {
			log.Fatalf("Failed to print version information: %#v", err)
		}

		fmt.Fprintln(os.Stdout, v)
		os.Exit(0)
	}

	// Initialize the secrets and selead secrets handler.
	secretsHandler := secrets.NewSecretHandler(*outputFormat)
	sealedSecretsHandler, err := secrets.NewSealedSecretsHandler(clientConfig, *certFile, *controllerNs, *controllerName, *outputFormat)
	if err != nil {
		log.Fatalf("Failed to initialize sealed secrets handler: %#v", err)
	}

	// Parse templates.
	indexTmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %#v", err)
	}

	listTmpl, err := template.ParseFiles("static/list.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %#v", err)
	}

	// Determine which example data should be shown
	example := yamlExample
	if *outputFormat == "json" {
		example = jsonExample
	}

	// Main handler.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Error  string
			Input  string
			Output string
			Mode   string
		}{
			"",
			example,
			"",
			*outputFormat,
		}

		if r.Method == http.MethodGet {
			urlValues := r.URL.Query()

			if urlValues.Get("namespace") != "" && urlValues.Get("name") != "" {
				secret, err := sealedSecretsHandler.GetSecret(urlValues.Get("namespace"), urlValues.Get("name"))
				if err != nil {
					log.Printf("Get secret error: %#v", err)
					data.Error = fmt.Sprintf("Get secret error: %s", err.Error())
				} else {
					data.Input = string(secret)
				}
			}

			indexTmpl.Execute(w, data)
			return
		} else if r.Method == http.MethodPost {
			data.Input = r.FormValue("input")

			if r.FormValue("submit") == "Encode" {
				// Encode a secret.
				secret, err := secretsHandler.Encode(data.Input)
				if err != nil {
					log.Printf("Encoding error: %#v", err)
					data.Error = fmt.Sprintf("Encoding error: %s", err.Error())
				} else {
					data.Input = string(secret)
				}
			} else if r.FormValue("submit") == "Decode" {
				// Decode a secret.
				secret, err := secretsHandler.Decode(data.Input)
				if err != nil {
					log.Printf("Decoding error: %#v", err)
					data.Error = fmt.Sprintf("Decoding error: %s", err.Error())
				} else {
					data.Input = string(secret)
				}
			} else if r.FormValue("submit") == "Seal" {
				// Create sealed secret.
				sealedsecret, err := sealedSecretsHandler.Seal([]byte(data.Input), scheme.Codecs)
				if err != nil {
					data.Error = fmt.Sprintf("Sealing error: %s", err.Error())
				} else {
					data.Output = string(sealedsecret)
				}
			}

			// Render template with the "calculated" input, output and error values.
			indexTmpl.Execute(w, data)
			return
		}

		data.Error = "Invalid method."
		indexTmpl.Execute(w, data)
		return
	})

	// Main handler.
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Error   string
			Secrets []secrets.Secret
		}{
			"",
			nil,
		}

		if r.Method == http.MethodGet {
			s, err := sealedSecretsHandler.List()
			if err != nil {
				log.Fatalf("Failed to list sealed secrets: %#v", err)
				data.Error = fmt.Sprintf("Failed to list sealed secrets: %s", err.Error())
			}

			data.Secrets = s

			listTmpl.Execute(w, data)
			return
		}

		data.Error = "Invalid method."
		listTmpl.Execute(w, data)
		return
	})

	// Health check handler.
	http.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		return
	})

	// Serve static assets.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Start the web interface.
	fmt.Printf("Starting \"Sealed Secrets\" Web Interface %s\n", version.Info())
	fmt.Printf("Build context %s\n", version.BuildContext())

	fmt.Printf("Listening on %s\n", *listenAddress)
	log.Println(http.ListenAndServe(*listenAddress, nil))
}
