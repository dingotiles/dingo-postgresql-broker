package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/mitchellh/mapstructure"
)

func (bkr *Broker) clusterFeaturesFromProvisionDetails(details brokerapi.ProvisionDetails) (features structs.ClusterFeatures, err error) {
	return bkr.clusterFeaturesFromParameters(details.Parameters)
}

func (bkr *Broker) clusterFeaturesFromUpdateDetails(details brokerapi.UpdateDetails) (features structs.ClusterFeatures, err error) {
	return bkr.clusterFeaturesFromParameters(details.Parameters)
}

func (bkr *Broker) clusterFeaturesFromParameters(params map[string]interface{}) (features structs.ClusterFeatures, err error) {
	err = mapstructure.Decode(params, &features)
	if err != nil {
		return
	}
	if features.NodeCount < 2 {
		features.NodeCount = 2
	}

	err = bkr.verifyClusterFeatures(features)
	return
}

func (bkr *Broker) verifyClusterFeatures(features structs.ClusterFeatures) (err error) {
	availableCells, err := bkr.scheduler.FilterCellsByGUIDs(features.CellGUIDsForNewNodes)
	if err != nil {
		return
	}
	if features.NodeCount < len(availableCells) {
		err = fmt.Errorf("Broker: Not enough Cell GUIDs (%v) for cluster of %d nodes", availableCells, features.NodeCount)
	}
	return
}
