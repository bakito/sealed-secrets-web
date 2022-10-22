package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/marshal"
	"gopkg.in/yaml.v3"
)

var (
	disableLoadSecrets            = flag.Bool("disable-load-secrets", false, "Disable the loading of existing secrets")
	enableWebLogs                 = flag.Bool("enable-web-logs", false, "Enable web logs")
	includeNamespaces             = flag.String("include-namespaces", "", "Optional space separated list if namespaces to be included in the sealed secret search")
	kubesealArgs                  = flag.String("kubeseal-arguments", "", "Deprecated use (sealed-secrets-service-name, sealed-secrets-service-namespace or sealed-secrets-cert-url)")
	sealedSecretsServiceName      = flag.String("sealed-secrets-service-name", "sealed-secrets", "Name of the sealed secrets service")
	sealedSecretsServiceNamespace = flag.String("sealed-secrets-service-namespace", "sealed-secrets", "Namespace of the sealed secrets service")
	sealedSecretsCertURL          = flag.String("sealed-secrets-cert-url", "", "URL sealed secrets certificate (required if sealed secrets is not reachable with in cluster service)")
	outputFormat                  = flag.String("format", "json", "Output format for sealed secret. Either json or yaml")
	initialSecretFile             = flag.String("initial-secret-file", "", "Define a file with the initial secret to be displayed. If empty, defaults are used.")
	webExternalURL                = flag.String("web-external-url", "", "Deprecated use (web-context)")
	webContext                    = flag.String("web-context", "", "The context the application is running on. (for example, if it is served via a reverse proxy)")
	printVersion                  = flag.Bool("version", false, "Print version information and exit")
	port                          = flag.Int("port", 8080, "Define the port to run the application on. (default: 8080)")
	config                        = flag.String("config", "", "Define the config file")
)

func Parse() (*Config, error) {
	flag.Parse()
	cfg := &Config{
		Web: Web{
			Port:    *port,
			Context: *webContext,
			Logger:  *enableWebLogs,
		},
		PrintVersion:       *printVersion,
		OutputFormat:       *outputFormat,
		DisableLoadSecrets: *disableLoadSecrets,
	}

	if *kubesealArgs != "" {
		fmt.Println("Argument 'kubeseal-arguments' is deprecated use (sealed-secrets-service-name, sealed-secrets-service-namespace or sealed-secrets-cert-url).")
	}
	if *webExternalURL != "" {
		fmt.Println("Argument 'web-external-url' is deprecated use (web-context).")
	}
	if *sealedSecretsServiceName != "" {
		cfg.SealedSecrets.Service = *sealedSecretsServiceName
	}
	if *sealedSecretsServiceNamespace != "" {
		cfg.SealedSecrets.Namespace = *sealedSecretsServiceNamespace
	}
	if *sealedSecretsCertURL != "" {
		cfg.SealedSecrets.CertURL = *sealedSecretsCertURL
	}
	if *includeNamespaces != "" {
		cfg.IncludeNamespaces = strings.Split(*includeNamespaces, " ")
	}
	if *initialSecretFile != "" {
		b, err := os.ReadFile(*initialSecretFile)
		if err != nil {
			return nil, err
		}
		cfg.InitialSecret = string(b)
	}

	if *config != "" {
		b, err := os.ReadFile(*config)
		if err != nil {
			return nil, err
		}

		if err = yaml.Unmarshal(b, cfg); err != nil {
			return nil, err
		}
	}

	if cfg.FieldFilter == nil {
		cfg.FieldFilter = &FieldFilter{
			Skip: [][]string{},
			SkipIfNil: [][]string{
				{"metadata", "creationTimestamp"},
				{"spec", "template", "data"},
				{"spec", "template", "metadata", "creationTimestamp"},
			},
		}
	}

	cfg.Marshaller = marshal.For(cfg.OutputFormat)

	if cfg.InitialSecret != "" {
		// Read and format the initial secret with the default marshaller
		sec := make(map[string]interface{})
		if err := cfg.Marshaller.Unmarshal([]byte(cfg.InitialSecret), sec); err != nil {
			return nil, fmt.Errorf("could not parse the initial secret: %w", err)
		}
	}

	return cfg, nil
}

type Config struct {
	Web                Web                `yaml:"web"`
	FieldFilter        *FieldFilter       `yaml:"fieldFilter,omitempty"`
	PrintVersion       bool               `yaml:"printVersion"`
	DisableLoadSecrets bool               `yaml:"disableLoadSecrets"`
	IncludeNamespaces  []string           `yaml:"includeNamespaces"`
	OutputFormat       string             `yaml:"outputFormat"`
	SealedSecrets      SealedSecrets      `yaml:"sealedSecrets"`
	InitialSecret      string             `yaml:"initialSecret"`
	Marshaller         marshal.Marshaller `yaml:"_"`
}

type Web struct {
	Port    int    `yaml:"port"`
	Context string `yaml:"context"`
	Logger  bool   `yaml:"logger"`
}

type SealedSecrets struct {
	Service   string `yaml:"service"`
	Namespace string `yaml:"namespace"`
	CertURL   string `yaml:"certURL,omitempty"`
}
