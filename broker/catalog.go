package broker

import "github.com/frodenas/brokerapi"

// Services is the catalog of services offered by the broker
func (broker *Broker) Services() brokerapi.CatalogResponse {
	return brokerapi.CatalogResponse{}
}

// LastOperation returns the status of the last operation on a service instance
func (broker *Broker) LastOperation(instanceID string) (brokerapi.LastOperationResponse, error) {
	return brokerapi.LastOperationResponse{}, nil
}
