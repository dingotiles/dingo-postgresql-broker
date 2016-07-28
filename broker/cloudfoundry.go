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

func (cf *CloudFoundryFromConfig) LookupServiceName(instanceID structs.ClusterID) (name string, err error) {
	var siResp serviceInstanceResponse
	r := cf.client.NewRequest("GET", fmt.Sprintf("/v2/service_instances/%d", instanceID))
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
