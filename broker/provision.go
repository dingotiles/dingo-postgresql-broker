package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	clusterInstance := state.NewClusterFromProvisionDetails(instanceID, details, bkr.etcdClient, bkr.config, bkr.logger)

	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, acceptsIncomplete)
	}

	logger := bkr.logger
	logger.Info("provision.start", lager.Data{})

	if bkr.state.ClusterExists(instanceID) {
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
				var restoredData *state.ClusterData
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
