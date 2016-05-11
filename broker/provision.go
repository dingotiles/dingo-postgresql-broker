package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	clusterInstance := cluster.NewClusterFromProvisionDetails(instanceID, details, bkr.etcdClient, bkr.config, bkr.logger)

	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, acceptsIncomplete)
	}

	logger := clusterInstance.Logger
	logger.Info("provision.start", lager.Data{})

	if clusterInstance.Exists() {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	canProvision := bkr.licenseCheck.CanProvision(details.ServiceID, details.PlanID)
	if !canProvision {
		return resp, false, fmt.Errorf("Quota for new service instances has been reached. Please contact Dingo Tiles to increase quota.")
	}

	clusterRequest := scheduler.NewRequest(clusterInstance, clusterInstance.Data.NodeCount, 20)

	go func() {
		err = clusterRequest.Perform()
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
				clusterInstance.TriggerClusterDataBackup(bkr.config.Callbacks)
				var restoredData *cluster.ClusterData
				err, restoredData = cluster.RestoreClusterDataBackup(clusterInstance.Data.InstanceID, bkr.config.Callbacks, logger)
				if err != nil || !restoredData.Equals(&clusterInstance.Data) {
					logger.Error("clusterdata.backup.failure", err, lager.Data{"clusterdata": clusterInstance.Data, "restoreddata": *restoredData})
					go func() {
						bkr.Deprovision(clusterInstance.Data.InstanceID, brokerapi.DeprovisionDetails{
							PlanID:    clusterInstance.Data.PlanID,
							ServiceID: clusterInstance.Data.ServiceID,
						}, true)
					}()
				}
			}

			logger.Info("provision.running.success", lager.Data{"cluster": clusterInstance.ClusterData()})
		}
	}()
	return resp, true, err
}
