package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Deprovision service instance
func (bkr *Broker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, acceptsIncomplete bool) (async bool, err error) {
	return bkr.deprovision(structs.ClusterID(instanceID), details, acceptsIncomplete)
}

func (bkr *Broker) deprovision(instanceID structs.ClusterID, details brokerapi.DeprovisionDetails, acceptsIncomplete bool) (async bool, err error) {
	logger := bkr.newLoggingSession("deprovision", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	if err = bkr.assertDeprovisionPrecondition(instanceID, details); err != nil {
		logger.Error("preconditions.error", err)
		return false, err
	}

	clusterState, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		logger.Error("load-cluster.error", err)
		return false, err
	}

	clusterModel := state.NewClusterStateModel(bkr.state, clusterState)
	err = bkr.scheduler.StopCluster(clusterModel)
	if err != nil {
		logger.Error("stop-cluster", err)
		return false, err
	}

	bkr.state.DeleteCluster(clusterModel.InstanceID())
	bkr.router.RemoveClusterAssignment(clusterModel.InstanceID())

	return false, nil
}

func (bkr *Broker) assertDeprovisionPrecondition(instanceID structs.ClusterID, details brokerapi.DeprovisionDetails) error {
	if bkr.state.ClusterExists(instanceID) == false {
		return fmt.Errorf("Service instance %s doesn't exist", instanceID)
	}

	if details.ServiceID == "" || details.PlanID == "" {
		return fmt.Errorf("API error - provide service_id and plan_id as URL parameters")
	}

	return nil
}
