package broker

import (
	"fmt"

	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)
	if cluster.Exists() {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	var nodeCount int
	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	} else {
		// 2-node default cluster
		nodeCount = 2
	}
	if nodeCount < 1 {
		return resp, false, fmt.Errorf("node-count parameter must be number greater than 0; preferrable 2 or more")
	}
	clusterRequest := servicechange.NewRequest(cluster, int(nodeCount), 20)

	clusterRequest.Perform()
	cluster.WaitForRoutingPortAllocation()
	return resp, false, nil
}
