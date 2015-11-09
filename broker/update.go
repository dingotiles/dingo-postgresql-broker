package broker

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Update service instance
func (bkr *Broker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, acceptsIncomplete bool) (async bool, err error) {
	details := brokerapi.ProvisionDetails{
		ServiceID:  updateDetails.ServiceID,
		PlanID:     updateDetails.PlanID,
		Parameters: updateDetails.Parameters,
	}

	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)
	err = cluster.Load()
	if err != nil {
		return false, err
	}

	var nodeCount int

	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	} else {
		nodeCount = int(cluster.NodeCount)
	}
	clusterRequest := servicechange.NewRequest(cluster, int(nodeCount), 20)
	clusterRequest.Perform()
	return false, nil
}
