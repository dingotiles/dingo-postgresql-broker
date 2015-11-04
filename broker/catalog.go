package broker

import "github.com/frodenas/brokerapi"

// Services is the catalog of services offered by the broker
func (broker *Broker) Services() brokerapi.CatalogResponse {
	return brokerapi.CatalogResponse{
		Services: []brokerapi.Service{
			{ID: "uuid", Name: "postgresql"},
		},
	}
}
