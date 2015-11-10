package config

import (
	"io/ioutil"

	"github.com/frodenas/brokerapi"
	"gopkg.in/yaml.v1"
)

// LoadServices catalog from a YAML file
func LoadServices(path string) (catalog *brokerapi.CatalogResponse, err error) {
	catalog = &brokerapi.CatalogResponse{}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(bytes, &catalog)
	if err != nil {
		return
	}

	return
}
