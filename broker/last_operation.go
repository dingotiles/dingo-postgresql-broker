package broker

import "github.com/frodenas/brokerapi"

// LastOperation returns the status of the last operation on a service instance
func (broker *Broker) LastOperation(instanceID string) (brokerapi.LastOperationResponse, error) {
	// Lookup /clusterrequest/instanceID
	return brokerapi.LastOperationResponse{
		State:       brokerapi.LastOperationInProgress,
		Description: "todo",
	}, nil
}
