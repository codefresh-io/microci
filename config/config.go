package config

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type bardConfigStep struct {
	Type string `yaml:"type,omitempty"`
	Desc string `yaml:"description,omitempty"`
}

type bardConfig struct {
	Version string                    `yaml:"version"`
	Steps   map[string]bardConfigStep `yaml:"steps"`
}

func unmarshal(data []byte) (*bardConfig, error) {
	var config bardConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	log.Debugf("Bard config = %+v\n", config)

	return &config, nil
}

func readConfigFile(configFile string) (*bardConfig, error) {
	log.Debugf("Bard config file = %s\n", configFile)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	return unmarshal(data)
}
