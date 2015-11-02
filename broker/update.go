package broker

import "github.com/frodenas/brokerapi"

// Update service instance
func (broker *Broker) Update(instanceID string, details brokerapi.UpdateDetails, acceptsIncomplete bool) (bool, error) {
	return true, nil
}
