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
	Path    string   `yaml:"path"`
	Target  string   `yaml:"target"`
	Headers []Header `yaml:"headers"`
}

type Server struct {
	Host     string    `yaml:"host"`
	Addr     string    `yaml:"addr"`
	Rewrites []Rewrite `yaml:"rewrite"`
}

func ParseRewritesFile(filename string) ([]Server, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	routes := make([]Server, 0)
	err = yaml.NewDecoder(f).Decode(&routes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return routes, nil
}
