package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Recreate service instance; invoked via Provision endpoint
func (bkr *Broker) Recreate(instanceID structs.ClusterID, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	logger := bkr.newLoggingSession("recreate", lager.Data{})
	defer logger.Info("stop")

	features, err := structs.ClusterFeaturesFromParameters(details.Parameters)
	if err != nil {
		logger.Error("cluster-features", err)
		return resp, false, err
	}

	if err = bkr.assertRecreatePrecondition(instanceID, features); err != nil {
		logger.Error("preconditions.error", err)
		return resp, false, err
	}

	recreationData, err := bkr.callbacks.RestoreRecreationData(instanceID)
	if err != nil {
		err = fmt.Errorf("Cannot recreate service from backup; unable to restore original service instance data: %s", err)
		return
	}

	clusterState := bkr.initClusterStateFromRecreationData(recreationData)
	clusterModel := state.NewClusterStateModel(bkr.state, clusterState)
	err = clusterModel.ResetClusterPlan()
	if err != nil {
		logger.Error("reset-cluster-plan", err)
		return
	}

	go func() {
		err := bkr.scheduler.RunCluster(clusterModel, features)
		if err != nil {
			clusterModel.PlanError(err)
			logger.Error("run-cluster", err)
			return
		}

		err = bkr.router.AssignPortToCluster(clusterModel.InstanceID(), clusterModel.AllocatedPort())
		if err != nil {
			clusterModel.PlanError(err)
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

func (bkr *Broker) assertRecreatePrecondition(instanceID structs.ClusterID, features structs.ClusterFeatures) error {
	if bkr.state.ClusterExists(instanceID) {
		return fmt.Errorf("service instance %s already exists", instanceID)
	}

	return bkr.scheduler.VerifyClusterFeatures(features)
}
