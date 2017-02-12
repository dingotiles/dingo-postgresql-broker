package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// APISpecification configures access to the central API
// Each are loaded from environment variables via APISpec()
// e.g. ClusterName with envconfig "cluster" is loaded from DINGO_CLUSTER
// e.g. NodeName with envconfig "node" is loaded from DINGO_NODE
type APISpecification struct {
	ImageVersion            string `required:"true" envconfig:"image_version"`
	ClusterName             string `required:"true" envconfig:"cluster"`
	NodeName                string `required:"false" envconfig:"node"`
	OrgAuthToken            string `required:"true" envconfig:"org_token"`
	APIURI                  string `required:"true" envconfig:"api_uri"`
	PatroniDefaultPath      string `default:"/patroni/patroni-default-values.yml" envconfig:"patroni_default_path"`
	PatroniPostgresStartCmd string `envconfig:"patroni_postgres_start_command"`
}

var apiSpec *APISpecification

// APISpec creates an APISpecification from environment variables prefixed with DINGO
// E.g. DINGO_IMAGE_VERSION=1.1 is placed in APISpecification{ImageVersion: "1.1"}
func APISpec() *APISpecification {
	if apiSpec == nil {
		apiSpec = &APISpecification{}
		err := envconfig.Process("dingo", apiSpec)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	return apiSpec
}
