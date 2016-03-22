package broker

import (
	"fmt"

	"github.com/dingotiles/patroni-broker/servicechange"
	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)

	logger := cluster.Logger
	logger.Info("provision.start", lager.Data{})

	if cluster.Exists() {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	// 1-node default cluster
	nodeCount := 1
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

	clusterRequest.Perform()
	cluster.WaitForRoutingPortAllocation()

	err = cluster.WaitForAllRunning()

	if err != nil {
		logger.Info("provision.end.with-error", lager.Data{"err": err})
	} else {
		logger.Info("provision.end.success", lager.Data{"cluster": cluster.ClusterData()})
		// provisionCallback := bkr.Config.Callbacks.ProvisionSuccess
		// if provisionCallback != nil {
		// 	logger.Info("callbacks.provision.running", lager.Data{"command": provisionCallback})
		// 	out, err := exec.Command(provisionCallback.Command, provisionCallback.Arguments...).CombinedOutput()
		// 	if err != nil {
		// 		logger.Error("callbacks.provision.error", err)
		// 	} else {
		// 		logger.Info("callbacks.provision.success", lager.Data{"output": out})
		// 	}
		// }
	}
	return resp, false, err
}
