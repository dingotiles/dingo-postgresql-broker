package broker

import "github.com/frodenas/brokerapi"

// Unbind to remove access to service instance
func (broker *Broker) Unbind(instanceID string, bindingID string, details brokerapi.UnbindDetails) error {
	return nil
}
