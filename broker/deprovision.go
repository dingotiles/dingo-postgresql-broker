package broker

import "github.com/frodenas/brokerapi"

// Deprovision service instance
func (broker *Broker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, acceptsIncomplete bool) (bool, error) {
	return true, nil
}
