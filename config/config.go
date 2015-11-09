package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

// Config is the brokers configuration
type Config struct {
	Broker   Broker     `yaml:"broker"`
	Backends []*Backend `yaml:"backends"`
	KVStore  KVStore    `yaml:"kvstore"`
}

// Broker connection configuration
type Broker struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Backend describes a configured set of backend brokers
// TODO dynamicly load from KV store
type Backend struct {
	GUID             string `yaml:"guid"`
	AvailabilityZone string `yaml:"availability_zone"`
	URI              string `yaml:"uri"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
}

// KVStore describes the KV store used by all the components
type KVStore struct {
	Type     string   `yaml:"type"`
	Machines []string `yaml:"machines"`
	Username string
	Password string
}

// LoadConfig from a YAML file
func LoadConfig(path string) (config *Config, err error) {
	config = &Config{}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return
	}

	// defaults
	if config.Broker.Username == "" {
		config.Broker.Username = "starkandwayne"
	}
	if config.Broker.Password == "" {
		config.Broker.Password = "starkandwayne"
	}
	if config.Broker.Port == 0 {
		config.Broker.Port = 3000
	}

	return
}
