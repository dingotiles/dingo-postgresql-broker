package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/servicechange"
	"github.com/dingotiles/dingo-postgresql-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	cluster := serviceinstance.NewClusterFromProvisionDetails(instanceID, details, bkr.etcdClient, bkr.config, bkr.logger)

	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, acceptsIncomplete)
	}

	logger := cluster.Logger
	logger.Info("provision.start", lager.Data{})

	if cluster.Exists() {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	canProvision := bkr.licenseCheck.CanProvision(details.ServiceID, details.PlanID)
	if !canProvision {
		return resp, false, fmt.Errorf("Quota for new service instances has been reached. Please contact Dingo Tiles to increase quota.")
	}

	// 2-node default cluster
	nodeCount := 2
	nodeSize := 20 // meaningless at moment
	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	}
	if nodeCount < 1 {
		logger.Info("provision.start.node-count-too-low", lager.Data{"node-count": nodeCount})
		nodeCount = 1
	}
	clusterRequest := servicechange.NewRequest(cluster, int(nodeCount), nodeSize)

	go func() {
		err = clusterRequest.Perform()
		if err != nil {
			logger.Error("provision.perform.error", err)
		}

		err = cluster.WaitForAllRunning()
		if err == nil {
			// if cluster is running, then wait until routing port operational
			err = cluster.WaitForRoutingPortAllocation()
		}

		if err != nil {
			logger.Error("provision.running.error", err)
		} else {

			if bkr.config.SupportsClusterDataBackup() {
				cluster.TriggerClusterDataBackup(bkr.config.Callbacks)
				var restoredData *serviceinstance.ClusterData
				err, restoredData = serviceinstance.RestoreClusterDataBackup(cluster.Data.InstanceID, bkr.config.Callbacks, logger)
				if err != nil || !restoredData.Equals(&cluster.Data) {
					logger.Error("clusterdata.backup.failure", err, lager.Data{"clusterdata": cluster.Data, "restoreddata": *restoredData})
					go func() {
						bkr.Deprovision(cluster.Data.InstanceID, brokerapi.DeprovisionDetails{
							PlanID:    cluster.Data.PlanID,
							ServiceID: cluster.Data.ServiceID,
						}, true)
					}()
				}
			}

			logger.Info("provision.running.success", lager.Data{"cluster": cluster.ClusterData()})
		}
	}()
	return resp, true, err
}
