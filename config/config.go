package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Rewrite struct {
	Path     string   `yaml:"path"`
	Target   string   `yaml:"target"`
	Redirect bool     `yaml:"redirect"`
	Headers  []Header `yaml:"headers"`
}

type Server struct {
	Host string `yaml:"host"`
	Addr string `yaml:"addr"`
	Ssl  bool   `yaml:"ssl"`
}

type Domain struct {
	Name     string    `yaml:"name"`
	Rewrites []Rewrite `yaml:"rewrites"`
	// TODO: headers
}

type Listener struct {
	Addr string `yaml:"addr"`
	SSL bool
	Domains []Domain
}

type ConfigFile struct {
	Listeners []Listener
}


func ParseRewritesFile(filename string) (*ConfigFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	routes := &ConfigFile{}
	err = yaml.NewDecoder(f).Decode(&routes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return routes, nil
}
