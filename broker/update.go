package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
)

// Update service instance
func (bkr *Broker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, acceptsIncomplete bool) (async bool, err error) {
	details := brokerapi.ProvisionDetails{
		ServiceID:  updateDetails.ServiceID,
		PlanID:     updateDetails.PlanID,
		Parameters: updateDetails.Parameters,
	}

	cluster, err := bkr.state.LoadCluster(instanceID)

	var nodeCount int
	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	} else {
		nodeCount = int(cluster.MetaData().TargetNodeCount)
	}
	if nodeCount < 1 {
		return false, fmt.Errorf("node-count parameter must be number greater than 0; preferrable 2 or more")
	}
	cluster.SetTargetNodeCount(nodeCount)
	clusterRequest := bkr.scheduler.NewRequest(cluster)
	bkr.scheduler.Execute(clusterRequest)
	return false, nil
}
