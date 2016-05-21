package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
)

func (bkr *Broker) clusterFeaturesFromProvisionDetails(details brokerapi.ProvisionDetails) structs.ClusterFeatures {
	targetNodeCount := defaultNodeCount
	if rawNodeCount := details.Parameters["node-count"]; rawNodeCount != nil {
		targetNodeCount = int(rawNodeCount.(float64))
	}
	return structs.ClusterFeatures{
		NodeCount: targetNodeCount,
	}
}

func (bkr *Broker) clusterFeaturesFromUpdateDetails(details brokerapi.UpdateDetails) structs.ClusterFeatures {
	targetNodeCount := defaultNodeCount
	if rawNodeCount := details.Parameters["node-count"]; rawNodeCount != nil {
		targetNodeCount = int(rawNodeCount.(float64))
	}
	return structs.ClusterFeatures{
		NodeCount: targetNodeCount,
	}
}
