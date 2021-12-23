package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/marshal"
	"gopkg.in/yaml.v3"
)

var (
	disableLoadSecrets = flag.Bool("disable-load-secrets", false, "Disable the loading of existing secrets")
	includeNamespaces  = flag.String("include-namespaces", "", "Optional space separated list if namespaces to be included in the sealed secret search")
	kubesealArgs       = flag.String("kubeseal-arguments", "", "Arguments which are passed to kubeseal")
	outputFormat       = flag.String("format", "json", "Output format for sealed secret. Either json or yaml")
	initialSecretFile  = flag.String("initial-secret-file", "", "Define a file with the initial secret to be displayed. If empty, defaults are used.")
	webExternalURL     = flag.String("web-external-url", "", "The URL under which the Sealed Secrets Web Interface is externally reachable (for example, if it is served via a reverse proxy).")
	printVersion       = flag.Bool("version", false, "Print version information and exit")
	port               = flag.Int("port", 8080, "Define the port to run the application on. (default: 8080)")
	config             = flag.String("config", "", "Define the config file")
)

func Parse() (*Config, error) {
	flag.Parse()
	cfg := &Config{
		Web: Web{
			Port:        *port,
			ExternalURL: *webExternalURL,
		},
		PrintVersion:       *printVersion,
		OutputFormat:       *outputFormat,
		DisableLoadSecrets: *disableLoadSecrets,
	}

	if *kubesealArgs != "" {
		cfg.KubesealArgs = strings.Split(*kubesealArgs, " ")
	}
	if *includeNamespaces != "" {
		cfg.IncludeNamespaces = strings.Split(*includeNamespaces, " ")
	}
	if *initialSecretFile != "" {
		b, err := ioutil.ReadFile(*initialSecretFile)
		if err != nil {
			return nil, err
		}
		cfg.InitialSecret = string(b)
	}

	if *config != "" {
		b, err := ioutil.ReadFile(*config)
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
		v, _ := cfg.Marshaller.Marshal(sec)
		cfg.InitialSecret = string(v)
	}

	for _, arg := range cfg.KubesealArgs {
		if strings.HasPrefix(arg, "--format") {
			return nil, fmt.Errorf("'--format' is not allowed as kubeseal argument")
		} else if strings.HasPrefix(arg, "-o") {
			return nil, fmt.Errorf("'-o' is not allowed as kubeseal argument")
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
	KubesealArgs       []string           `yaml:"kubesealArgs"`
	InitialSecret      string             `yaml:"initialSecret"`
	Marshaller         marshal.Marshaller `yaml:"_"`
}

type Web struct {
	Port        int    `yaml:"port"`
	ExternalURL string `yaml:"externalUrl"`
}
