package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the ratproxy.yaml.
type Config struct {
	Listen  string    `yaml:"listen"`
	UseTLS  bool      `yaml:"useTLS"`
	Targets []*Target `yaml:"targets"`
}

const (
	TargetTypeLambda string = "lambda"
	TargetTypeEmpty  string = ""
)

// Target is a proxy target based on a path.
type Target struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Path   string `yaml:"path"`
	Type   string `yaml:"type"`
}

// NewConfig reads the yaml configuration file.
func NewConfig() (*Config, error) {
	configFilename := os.Getenv("RATPROXY_FILENAME")
	if configFilename == "" {
		configFilename = "ratproxy.yaml"
	}

	log.Printf("ratproxy: reading config at %v\n", configFilename)
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
