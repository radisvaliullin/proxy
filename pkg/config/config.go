package config

import (
	"log"
	"os"

	"github.com/radisvaliullin/proxy/pkg/proxy"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy proxy.Config `yaml:"proxy"`
}

func New() (Config, error) {

	c := Config{}

	configBytes, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		log.Printf("config: read file: %v", err)
		return c, err
	}

	if err := yaml.Unmarshal(configBytes, &c); err != nil {
		log.Printf("config: unmarshal config yaml: %v", err)
		return c, err
	}

	return c, nil
}
