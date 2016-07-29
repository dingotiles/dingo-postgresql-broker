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
	Broker       Broker                  `yaml:"broker"`
	Cells        []*Cell                 `yaml:"cells"`
	Etcd         Etcd                    `yaml:"etcd"`
	Callbacks    Callbacks               `yaml:"callbacks"`
	Catalog      brokerapi.Catalog       `yaml:"catalog"`
	Scheduler    Scheduler               `yaml:"scheduler"`
	CloudFoundry CloudFoundryCredentials `yaml:"cf"`
}

func (cfg *Config) SupportsClusterDataBackup() bool {
	return cfg.Callbacks.ClusterDataBackup != nil && cfg.Callbacks.ClusterDataRestore != nil
}

// Broker connection configuration
type Broker struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	BindHost string `yaml:"bind_host"`
}

type Scheduler struct {
	Cells []*Cell
	Etcd  Etcd
}

// Cell describes a configured set of cell brokers
// TODO dynamicly load from KV store
type Cell struct {
	GUID             string `yaml:"guid"`
	AvailabilityZone string `yaml:"availability_zone"`
	URI              string `yaml:"uri"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
}

// KVStore describes the KV store used by all the components
type Etcd struct {
	Machines []string `yaml:"machines"`
}

// CloudFoundryCredentials describes credentials for looking up service instance info/name
// Requires SpaceDeveloper access to all Spaces for which access is enabled.
type CloudFoundryCredentials struct {
	ApiAddress        string `yaml:"api_url"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	SkipSslValidation bool   `yaml:"skip_ssl_validation"`
}

// Callbacks allows plug'n'play scripts to be run when events have completed
type Callbacks struct {
	ClusterDataBackup  *CallbackCommand `yaml:"clusterdata_backup"`
	ClusterDataRestore *CallbackCommand `yaml:"clusterdata_restore"`
}

// CallbackCommand describes a command that can be run via os/exec's Command
type CallbackCommand struct {
	Command   string   `yaml:"cmd"`
	Arguments []string `yaml:"args"`
}

// LoadConfig from a YAML file
func LoadConfig(path string) (cfg *Config, err error) {
	cfg = &Config{}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return
	}

	// defaults
	if cfg.Broker.Username == "" {
		cfg.Broker.Username = "starkandwayne"
	}
	if cfg.Broker.Password == "" {
		cfg.Broker.Password = "starkandwayne"
	}
	if cfg.Broker.Port == 0 {
		cfg.Broker.Port = 3000
	}

	for _, cell := range cfg.Cells {
		match, err := regexp.MatchString("^http", cell.URI)
		if !match || err != nil {
			cell.URI = fmt.Sprintf("http://%s", cell.URI)
		}
	}

	cfg.Scheduler = Scheduler{
		Etcd:  cfg.Etcd,
		Cells: cfg.Cells,
	}

	return
}
