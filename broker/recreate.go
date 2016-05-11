package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Recreate(instanceID string, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	logger := bkr.logger.Session("recreate", lager.Data{
		"instance-id": instanceID,
	})

	logger.Info("start", lager.Data{})
	var clusterdata *cluster.ClusterData
	err, clusterdata = cluster.RestoreClusterDataBackup(instanceID, bkr.config.Callbacks, bkr.logger)
	if err != nil {
		err = fmt.Errorf("Cannot recreate service from backup; unable to restore original service instance data: %s", err)
		return
	}

	cluster := cluster.NewClusterFromRestoredData(instanceID, clusterdata, bkr.etcdClient, bkr.config, logger)
	logger = bkr.logger
	logger.Info("me", lager.Data{"cluster": cluster})

	if cluster.Exists() {
		logger.Info("exists")
		err = fmt.Errorf("Service instance %s still exists in etcd, please clean it out before recreating cluster", instanceID)
		return
	} else {
		logger.Info("not-exists")
	}

	// Restore port allocation from cluster.Data
	key := fmt.Sprintf("/routing/allocation/%s", cluster.Data.InstanceID)
	_, err = bkr.etcdClient.Set(key, cluster.Data.AllocatedPort, 0)
	if err != nil {
		logger.Error("routing-allocation.error", err)
		return
	}
	logger.Info("routing-allocation.restored", lager.Data{"allocated-port": cluster.Data.AllocatedPort})

	nodeCount := cluster.Data.NodeCount
	if nodeCount < 1 {
		nodeCount = 1
	}
	cluster.Data.NodeCount = 0
	clusterRequest := scheduler.NewRequest(cluster, nodeCount, logger)
	err = clusterRequest.Perform()
	if err != nil {
		logger.Error("provision.perform.error", err)
		return resp, false, err
	}

	// if port is allocated, then wait to confirm containers are running
	err = cluster.WaitForAllRunning()

	if err != nil {
		logger.Error("provision.running.error", err)
		return
	}
	logger.Info("provision.running.success", lager.Data{"cluster": cluster.Data})
	cluster.TriggerClusterDataBackup(bkr.config.Callbacks)
	return
}
