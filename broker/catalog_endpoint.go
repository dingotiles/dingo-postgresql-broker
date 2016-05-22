package broker

import "github.com/frodenas/brokerapi"

// Services is the catalog of services offered by the broker
func (bkr *Broker) Services() brokerapi.CatalogResponse {
	result := brokerapi.CatalogResponse{}
	result.Services = bkr.config.Catalog.Services

	return result
}
