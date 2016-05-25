package broker

import (
	"fmt"
	"reflect"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

const defaultNodeCount = 2

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, details, acceptsIncomplete)
	}

	logger := bkr.newLoggingSession("provision", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	if err = bkr.assertProvisionPrecondition(instanceID, details); err != nil {
		logger.Error("preconditions.error", err)
		return resp, false, err
	}

	port, err := bkr.router.AllocatePort()
	clusterState := bkr.initCluster(instanceID, port, details)

	if bkr.callbacks.Configured() {
		bkr.callbacks.WriteRecreationData(clusterState.RecreationData())
		data, err := bkr.callbacks.RestoreRecreationData(clusterState.InstanceID)
		if !reflect.DeepEqual(clusterState.RecreationData(), data) {
			logger.Error("recreation-data.failure", err)
			return resp, false, err
		}
	}

	go func() {
		features := bkr.clusterFeaturesFromProvisionDetails(details)
		scheduledCluster, err := bkr.scheduler.RunCluster(clusterState, features)
		if err != nil {
			logger.Error("run-cluster", err)
		}

		err = bkr.router.AssignPortToCluster(scheduledCluster.InstanceID, port)
		if err != nil {
			logger.Error("assign-port", err)
		}

		err = bkr.state.SaveCluster(scheduledCluster)
		if err != nil {
			logger.Error("assign-port", err)
		}
	}()
	return resp, true, err
}

func (bkr *Broker) initCluster(instanceID string, port int, details brokerapi.ProvisionDetails) structs.ClusterState {
	return structs.ClusterState{
		InstanceID:       instanceID,
		OrganizationGUID: details.OrganizationGUID,
		PlanID:           details.PlanID,
		ServiceID:        details.ServiceID,
		SpaceGUID:        details.SpaceGUID,
		AllocatedPort:    port,
		AdminCredentials: structs.PostgresCredentials{
			Username: "pgadmin",
			Password: NewPassword(16),
		},
		SuperuserCredentials: structs.PostgresCredentials{
			Username: "postgres",
			Password: NewPassword(16),
		},
		AppCredentials: structs.PostgresCredentials{
			Username: "appuser",
			Password: NewPassword(16),
		},
	}
}

func (bkr *Broker) assertProvisionPrecondition(instanceID string, details brokerapi.ProvisionDetails) error {
	if bkr.state.ClusterExists(instanceID) {
		return fmt.Errorf("service instance %s already exists", instanceID)
	}

	canProvision := bkr.licenseCheck.CanProvision(details.ServiceID, details.PlanID)
	if !canProvision {
		return fmt.Errorf("Quota for new service instances has been reached. Please contact Dingo Tiles to increase quota.")
	}

	return nil
}
