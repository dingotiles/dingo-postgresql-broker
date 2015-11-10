package config

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/frodenas/brokerapi"
	"gopkg.in/yaml.v1"
)

// Config is the brokers configuration
type Config struct {
	Broker   Broker                    `yaml:"broker"`
	Router   Router                    `yaml:"router"`
	Backends []*Backend                `yaml:"backends"`
	KVStore  KVStore                   `yaml:"kvstore"`
	Catalog  brokerapi.CatalogResponse `yaml:"catalog"`
}

// Broker connection configuration
type Broker struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Router advertising info
type Router struct {
	Hostname string `yaml:"hostname"`
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

type Catalog struct {
	Services []Service
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

	for _, backend := range config.Backends {
		match, err := regexp.MatchString("^http", backend.URI)
		if !match || err != nil {
			backend.URI = fmt.Sprintf("http://%s", backend.URI)
		}
	}

	return
}
