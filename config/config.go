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
	Broker         Broker            `yaml:"broker"`
	Scheduler      Scheduler         `yaml:"scheduler"`
	Etcd           Etcd              `yaml:"etcd"`
	Callbacks      Callbacks         `yaml:"callbacks"`
	Catalog        brokerapi.Catalog `yaml:"catalog"`
	LicenseText    string            `yaml:"license_text"`
	LicenseDetails *LicenseDetails
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
	Backends []*Backend `yaml:"backends"`
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
type Etcd struct {
	Machines []string `yaml:"machines"`
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

	for _, backend := range cfg.Scheduler.Backends {
		match, err := regexp.MatchString("^http", backend.URI)
		if !match || err != nil {
			backend.URI = fmt.Sprintf("http://%s", backend.URI)
		}
	}

	cfg.LicenseDetails, err = NewLicenseDetailsFromLicenseText(cfg.LicenseText)
	if err != nil {
		fmt.Println(err)
		err = nil // its not that bad of an error at this stage
	} else {
		fmt.Printf("License decoded for %s, plans %#v\n", cfg.LicenseDetails.CompanyName, cfg.LicenseDetails.Plans)
	}

	return
}
