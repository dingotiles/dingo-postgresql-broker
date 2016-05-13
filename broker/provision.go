package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

const defaultNodeCount = 2

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, acceptsIncomplete)
	}

	logger := bkr.logger
	logger.Info("provision.start")

	if bkr.state.ClusterExists(instanceID) {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	canProvision := bkr.licenseCheck.CanProvision(details.ServiceID, details.PlanID)
	if !canProvision {
		return resp, false, fmt.Errorf("Quota for new service instances has been reached. Please contact Dingo Tiles to increase quota.")
	}

	clusterInstance, err := bkr.state.InitializeCluster(bkr.initClusterData(instanceID, details))
	if err != nil {
		logger.Error("provision.error", err)
		return resp, false, fmt.Errorf("Could not provision service instance. Error: %v", err)
	}
	clusterRequest := bkr.scheduler.NewRequest(clusterInstance, clusterInstance.MetaData().TargetNodeCount)

	go func() {
		err = bkr.scheduler.Execute(clusterRequest)
		if err != nil {
			logger.Error("provision.perform.error", err)
		}

		err = clusterInstance.WaitForAllRunning()
		if err == nil {
			// if cluster is running, then wait until routing port operational
			err = clusterInstance.WaitForRoutingPortAllocation()
		}

		if err != nil {
			logger.Error("provision.running.error", err)
		} else {

			if bkr.config.SupportsClusterDataBackup() {
				state.TriggerClusterDataBackup(clusterInstance.MetaData(), bkr.config.Callbacks, logger)
				var restoredData *structs.ClusterData
				err, restoredData = state.RestoreClusterDataBackup(clusterInstance.MetaData().InstanceID, bkr.config.Callbacks, logger)
				metaData := clusterInstance.MetaData()
				if err != nil || !restoredData.Equals(&metaData) {
					logger.Error("clusterdata.backup.failure", err, lager.Data{"clusterdata": clusterInstance.MetaData(), "restoreddata": *restoredData})
					go func() {
						bkr.Deprovision(clusterInstance.MetaData().InstanceID, brokerapi.DeprovisionDetails{
							PlanID:    clusterInstance.MetaData().PlanID,
							ServiceID: clusterInstance.MetaData().ServiceID,
						}, true)
					}()
				}
			}

			logger.Info("provision.running.success", lager.Data{"cluster": clusterInstance.MetaData()})
		}
	}()
	return resp, true, err
}

func (bkr *Broker) initClusterData(instanceID string, details brokerapi.ProvisionDetails) *structs.ClusterData {
	targetNodeCount := defaultNodeCount
	if rawNodeCount := details.Parameters["node-count"]; rawNodeCount != nil {
		targetNodeCount = int(rawNodeCount.(float64))
	}
	return &structs.ClusterData{
		InstanceID:       instanceID,
		OrganizationGUID: details.OrganizationGUID,
		PlanID:           details.PlanID,
		ServiceID:        details.ServiceID,
		SpaceGUID:        details.SpaceGUID,
		TargetNodeCount:  targetNodeCount,
		AdminCredentials: structs.AdminCredentials{
			Username: "pgadmin",
			Password: NewPassword(16),
		},
	}
}
