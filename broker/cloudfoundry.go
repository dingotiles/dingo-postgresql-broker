package broker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/hashicorp/errwrap"
	"github.com/pivotal-golang/lager"
)

// CloudFoundryFromConfig wraps CF credentials
type CloudFoundryFromConfig struct {
	creds config.CloudFoundryCredentials
}

// NewCloudFoundryFromConfig creates wrapper fresh cfclient.Client on request
func NewCloudFoundryFromConfig(creds config.CloudFoundryCredentials, logger lager.Logger) (cf *CloudFoundryFromConfig, err error) {
	cf = &CloudFoundryFromConfig{}
	if creds.ApiAddress == "" {
		return cf, fmt.Errorf("Cloud Foundry credentials not provided")
	}
	cf.creds = creds
	return
}

// Client creates a new cfclient.Client that will generate fresh auth tokens
func (cf *CloudFoundryFromConfig) Client() (client *cfclient.Client, err error) {
	client, err = cfclient.NewClient(&cfclient.Config{
		ApiAddress:        cf.creds.ApiAddress,
		Username:          cf.creds.Username,
		Password:          cf.creds.Password,
		SkipSslValidation: cf.creds.SkipSslValidation,
	})
	return
}

// LookupServiceName will look up the user-provided service instance name
func (cf *CloudFoundryFromConfig) LookupServiceName(instanceID structs.ClusterID) (name string, err error) {
	client, err := cf.Client()
	if err != nil {
		return "", errwrap.Wrapf("Cannot lookup Service Name: {{err}}", err)
	}
	var siResp serviceInstanceResponse
	r := client.NewRequest("GET", fmt.Sprintf("/v2/service_instances/%s", instanceID))
	resp, err := client.DoRequest(r)
	if err != nil {
		return "", errwrap.Wrapf("Error querying service instances: {{err}}", err)
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errwrap.Wrapf("Error reading service instances response: {{err}}", err)
	}

	err = json.Unmarshal(resBody, &siResp)
	if err != nil {
		return "", errwrap.Wrapf("Error unmarshaling service instances response: {{err}}", err)
	}
	return siResp.Entity.Name, nil
}

type serviceInstanceResponse struct {
	Entity struct {
		Name string `json:"name"`
	} `json:"entity"`
}
