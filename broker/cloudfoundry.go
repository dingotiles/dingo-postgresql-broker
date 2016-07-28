package broker

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type CloudFoundryFromConfig struct {
	client *cfclient.Client
}

func NewCloudFoundryFromConfig(creds config.CloudFoundryCredentials, logger lager.Logger) (cf *CloudFoundryFromConfig, err error) {
	if creds.ApiAddress == "" {
		return nil, fmt.Errorf("Cloud Foundry credentials not provided")
	}
	client, err := cfclient.NewClient(&cfclient.Config{
		ApiAddress:        creds.ApiAddress,
		Username:          creds.Username,
		Password:          creds.Password,
		SkipSslValidation: creds.SkipSslValidation,
	})
	if err != nil {
		return
	}
	cf = &CloudFoundryFromConfig{client: client}
	return
}

func (cf *CloudFoundryFromConfig) LookupServiceName(structs.ClusterID) (serviceInstanceName string, err error) {
	return
}
