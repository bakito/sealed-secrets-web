package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func Parse() (*Config, error) {
	return parse(newFlags())
}

func parse(f *flags) (*Config, error) {
	flag.Parse()
	cfg := &Config{
		Web: Web{
			Port:    *f.port,
			Context: *f.webContext,
			Logger:  *f.enableWebLogs,
		},
		PrintVersion:       *f.printVersion,
		DisableLoadSecrets: *f.disableLoadSecrets,
	}

	if *f.kubesealArgs != "" {
		log.Println("Argument 'kubeseal-arguments' is deprecated use (sealed-secrets-service-name, sealed-secrets-service-namespace or sealed-secrets-cert-url).")
	}
	if *f.webExternalURL != "" {
		log.Println("Argument 'web-external-url' is deprecated use (web-context).")
	}

	if *f.sealedSecretsCertURL != "" {
		cfg.SealedSecrets.CertURL = *f.sealedSecretsCertURL
	} else {
		if *f.sealedSecretsServiceName != "" {
			cfg.SealedSecrets.Service = *f.sealedSecretsServiceName
		}
		if *f.sealedSecretsServiceNamespace != "" {
			cfg.SealedSecrets.Namespace = *f.sealedSecretsServiceNamespace
		}
	}
	if *f.includeNamespaces != "" {
		cfg.IncludeNamespaces = strings.Split(*f.includeNamespaces, " ")
	}
	if *f.initialSecretFile != "" {
		b, err := os.ReadFile(*f.initialSecretFile)
		if err != nil {
			return nil, err
		}
		cfg.InitialSecret = string(b)
	}

	if *f.config != "" {
		b, err := os.ReadFile(*f.config)
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

	cfg.Web.Context = sanitizeWebContext(cfg)

	cfg.Ctx = context.Background()

	return cfg, nil
}

func sanitizeWebContext(cfg *Config) string {
	wc := cfg.Web.Context
	if !strings.HasPrefix(wc, "/") &&
		!strings.HasPrefix(wc, "http://") &&
		!strings.HasPrefix(wc, "https://") {
		wc = "/" + wc
	}
	if !strings.HasSuffix(wc, "/") {
		wc = wc + "/"
	}
	return wc
}

type Config struct {
	Web                Web             `yaml:"web"`
	FieldFilter        *FieldFilter    `yaml:"fieldFilter,omitempty"`
	PrintVersion       bool            `yaml:"printVersion"`
	DisableLoadSecrets bool            `yaml:"disableLoadSecrets"`
	IncludeNamespaces  []string        `yaml:"includeNamespaces"`
	SealedSecrets      SealedSecrets   `yaml:"sealedSecrets"`
	InitialSecret      string          `yaml:"initialSecret"`
	Ctx                context.Context `yaml:"-"`
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

func (ss SealedSecrets) String() string {
	if ss.CertURL != "" {
		return fmt.Sprintf("Cert URL: %s", ss.CertURL)
	}
	return fmt.Sprintf("Namespace: %s / ServiceName: %s", ss.Namespace, ss.Service)
}

type flags struct {
	disableLoadSecrets            *bool
	enableWebLogs                 *bool
	includeNamespaces             *string
	kubesealArgs                  *string
	sealedSecretsServiceName      *string
	port                          *int
	config                        *string
	printVersion                  *bool
	webContext                    *string
	webExternalURL                *string
	initialSecretFile             *string
	sealedSecretsCertURL          *string
	sealedSecretsServiceNamespace *string
}

func newFlags() *flags {
	return &flags{
		disableLoadSecrets:            flag.Bool("disable-load-secrets", false, "Disable the loading of existing secrets"),
		enableWebLogs:                 flag.Bool("enable-web-logs", false, "Enable web logs"),
		includeNamespaces:             flag.String("include-namespaces", "", "Optional space separated list if namespaces to be included in the sealed secret search"),
		kubesealArgs:                  flag.String("kubeseal-arguments", "", "Deprecated use (sealed-secrets-service-name, sealed-secrets-service-namespace or sealed-secrets-cert-url)"),
		sealedSecretsServiceName:      flag.String("sealed-secrets-service-name", "sealed-secrets", "Name of the sealed secrets service"),
		sealedSecretsServiceNamespace: flag.String("sealed-secrets-service-namespace", "sealed-secrets", "Namespace of the sealed secrets service"),
		sealedSecretsCertURL:          flag.String("sealed-secrets-cert-url", "", "URL sealed secrets certificate (required if sealed secrets is not reachable with in cluster service)"),
		initialSecretFile:             flag.String("initial-secret-file", "", "Define a file with the initial secret to be displayed. If empty, defaults are used."),
		webExternalURL:                flag.String("web-external-url", "", "Deprecated use (web-context)"),
		webContext:                    flag.String("web-context", "/", "The context the application is running on. (for example, if it is served via a reverse proxy)"),
		printVersion:                  flag.Bool("version", false, "Print version information and exit"),
		port:                          flag.Int("port", 8080, "Define the port to run the application on. (default: 8080)"),
		config:                        flag.String("config", "", "Define the config file"),
	}
}
