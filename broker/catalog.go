package broker

import (
	"github.com/frodenas/brokerapi"
	"github.com/mitchellh/mapstructure"
)

// Services is the catalog of services offered by the broker
func (bkr *Broker) Services() brokerapi.CatalogResponse {
	result := &brokerapi.CatalogResponse{}
	err := mapstructure.Decode(&bkr.Config.Catalog, &result)
	if err != nil {
		bkr.Logger.Error("services.decode", err)
	}

	return *result
}
