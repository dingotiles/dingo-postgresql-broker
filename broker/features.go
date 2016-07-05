package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/mitchellh/mapstructure"
)

func (bkr *Broker) clusterFeaturesFromProvisionDetails(details brokerapi.ProvisionDetails) (features structs.ClusterFeatures, err error) {
	err = mapstructure.Decode(details.Parameters, &features)
	return
}

func (bkr *Broker) clusterFeaturesFromUpdateDetails(details brokerapi.UpdateDetails) (features structs.ClusterFeatures, err error) {
	err = mapstructure.Decode(details.Parameters, &features)
	return
}
