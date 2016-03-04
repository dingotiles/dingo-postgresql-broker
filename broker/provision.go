package broker

import (
	"fmt"

	"github.com/dingotiles/patroni-broker/servicechange"
	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)
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
		return resp, false, fmt.Errorf("node-count parameter must be number greater than 0; preferrable 2 or more")
	}
	clusterRequest := servicechange.NewRequest(cluster, int(nodeCount), nodeSize)

	clusterRequest.Perform()
	cluster.WaitForRoutingPortAllocation()

	err = cluster.WaitForAllRunning()
	return resp, false, err
}
