package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	clusterInstance := cluster.NewClusterFromProvisionDetails(instanceID, details, bkr.etcdClient, bkr.config, bkr.logger)

	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, acceptsIncomplete)
	}

	logger := bkr.logger
	logger.Info("provision.start", lager.Data{})

	if clusterInstance.Exists() {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	canProvision := bkr.licenseCheck.CanProvision(details.ServiceID, details.PlanID)
	if !canProvision {
		return resp, false, fmt.Errorf("Quota for new service instances has been reached. Please contact Dingo Tiles to increase quota.")
	}

	// 2-node default cluster
	nodeCount := 2
	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	}
	if nodeCount < 1 {
		logger.Info("provision.start.node-count-too-low", lager.Data{"node-count": nodeCount})
		nodeCount = 1
	}
	clusterRequest := bkr.scheduler.NewRequest(clusterInstance, int(nodeCount))

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

			logger.Info("provision.running.success", lager.Data{"cluster": clusterInstance.Data})
		}
	}()
	return resp, true, err
}
