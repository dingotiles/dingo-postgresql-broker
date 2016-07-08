package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Update service instance
func (bkr *Broker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, acceptsIncomplete bool) (async bool, err error) {
	logger := bkr.newLoggingSession("update", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	features, err := bkr.clusterFeaturesFromUpdateDetails(updateDetails)
	if err != nil {
		logger.Error("cluster-features", err)
		return false, err
	}

	if err := bkr.assertUpdatePrecondition(instanceID, features); err != nil {
		logger.Error("preconditions.error", err)
		return false, err
	}

	cluster, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		logger.Error("load-cluster.error", err)
		return false, err
	}

	go func() {
		schedulerCluster, err := bkr.scheduler.RunCluster(cluster, features)
		if err != nil {
			logger.Error("run-cluster", err)
		}

		err = bkr.state.SaveCluster(schedulerCluster)
		if err != nil {
			logger.Error("assign-port", err)
		}
	}()
	return true, err
}

func (bkr *Broker) assertUpdatePrecondition(instanceID string, features structs.ClusterFeatures) error {
	if bkr.state.ClusterExists(instanceID) == false {
		return fmt.Errorf("Service instance %s doesn't exist", instanceID)
	}

	return bkr.scheduler.VerifyClusterFeatures(features)
}
