package broker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type CloudFoundryFromConfig struct {
	client *cfclient.Client
}

func NewCloudFoundryFromConfig(creds config.CloudFoundryCredentials, logger lager.Logger) (cf *CloudFoundryFromConfig, err error) {
	cf = &CloudFoundryFromConfig{}
	if creds.ApiAddress == "" {
		return cf, fmt.Errorf("Cloud Foundry credentials not provided")
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
	cf.client = client
	return
}

func (cf *CloudFoundryFromConfig) LookupServiceName(instanceID structs.ClusterID) (name string, err error) {
	if cf.client == nil {
		return "", fmt.Errorf("Cannot lookup Service Name for %d without Cloud Foundry credentials", instanceID)
	}
	var siResp serviceInstanceResponse
	r := cf.client.NewRequest("GET", fmt.Sprintf("/v2/service_instances/%s", instanceID))
	resp, err := cf.client.DoRequest(r)
	if err != nil {
		return "", fmt.Errorf("Error querying service instances %v", err)
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading service instances response: %v", err)
	}

	err = json.Unmarshal(resBody, &siResp)
	if err != nil {
		return "", fmt.Errorf("Error unmarshaling service instances response %v", err)
	}
	return siResp.Entity.Name, nil
}

type serviceInstanceResponse struct {
	Entity struct {
		Name string `json:"name"`
	} `json:"entity"`
}
