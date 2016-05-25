package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Recreate(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	logger := bkr.newLoggingSession("recreate", lager.Data{})
	defer logger.Info("stop")

	if err = bkr.assertRecreatePrecondition(instanceID); err != nil {
		logger.Error("preconditions.error", err)
		return resp, false, err
	}

	recreationData, err := bkr.callbacks.RestoreRecreationData(instanceID)
	if err != nil {
		err = fmt.Errorf("Cannot recreate service from backup; unable to restore original service instance data: %s", err)
		return
	}

	clusterState := bkr.initClusterStateFromRecreationData(recreationData)
	go func() {
		features := bkr.clusterFeaturesFromProvisionDetails(details)
		scheduledCluster, err := bkr.scheduler.RunCluster(clusterState, features)
		if err != nil {
			logger.Error("run-cluster", err)
		}

		err = bkr.state.SaveCluster(scheduledCluster)
		if err != nil {
			logger.Error("save-cluster", err)
		}

		err = bkr.router.AssignPortToCluster(scheduledCluster.InstanceID, scheduledCluster.AllocatedPort)
		if err != nil {
			logger.Error("assign-port", err)
		}
	}()

	return resp, true, err
}

func (bkr *Broker) initClusterStateFromRecreationData(recreationData *structs.ClusterRecreationData) structs.ClusterState {
	return structs.ClusterState{
		InstanceID:           recreationData.InstanceID,
		ServiceID:            recreationData.ServiceID,
		PlanID:               recreationData.PlanID,
		OrganizationGUID:     recreationData.OrganizationGUID,
		SpaceGUID:            recreationData.SpaceGUID,
		AdminCredentials:     recreationData.AdminCredentials,
		AppCredentials:       recreationData.AppCredentials,
		SuperuserCredentials: recreationData.SuperuserCredentials,
		AllocatedPort:        recreationData.AllocatedPort,
	}
}

func (bkr *Broker) assertRecreatePrecondition(instanceID string) error {
	if bkr.state.ClusterExists(instanceID) {
		return fmt.Errorf("service instance %s already exists", instanceID)
	}

	return nil
}
