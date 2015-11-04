package broker

import "github.com/frodenas/brokerapi"

// Bind returns access credentials for a service instance
func (broker *Broker) Bind(instanceID string, bindingID string, details brokerapi.BindDetails) (brokerapi.BindingResponse, error) {
	return brokerapi.BindingResponse{
		Credentials: brokerapi.CredentialsHash{
			Host: "10.10.10.10",
			Port: 40000,
		},
	}, nil
}
