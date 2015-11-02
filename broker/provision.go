package broker

import "github.com/frodenas/brokerapi"

// Provision a new service instance
func (broker *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (brokerapi.ProvisioningResponse, bool, error) {
	return brokerapi.ProvisioningResponse{}, true, nil
}
