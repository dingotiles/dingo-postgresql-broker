package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/mitchellh/mapstructure"
)

// Services is the catalog of services offered by the broker
func (bkr *Broker) Services() brokerapi.CatalogResponse {
	result := &brokerapi.CatalogResponse{}
	fmt.Printf("%#v\n", bkr.Config.Catalog)
	fmt.Printf("%#v\n", bkr.Config.Catalog.Services[0].Metadata)
	fmt.Printf("%#v\n", bkr.Config.Catalog.Services[0].Plans[0].Metadata)
	err := mapstructure.Decode(&bkr.Config.Catalog, &result)
	if err != nil {
		bkr.Logger.Error("services.decode", err)
	}

	return *result
}
