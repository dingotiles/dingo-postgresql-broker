package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

// Config is the brokers configuration
type Config struct {
	Backends []*Backend `yaml:"backends"`
	KVStore  KVStore    `yaml:"kvstore"`
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
	return
}
