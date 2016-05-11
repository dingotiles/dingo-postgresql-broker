package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
)

// Update service instance
func (bkr *Broker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, acceptsIncomplete bool) (async bool, err error) {
	details := brokerapi.ProvisionDetails{
		ServiceID:  updateDetails.ServiceID,
		PlanID:     updateDetails.PlanID,
		Parameters: updateDetails.Parameters,
	}

	cluster := state.NewClusterFromProvisionDetails(instanceID, details, bkr.etcdClient, bkr.config, bkr.logger)
	err = cluster.Load()
	if err != nil {
		return false, err
	}

	var nodeCount int
	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	} else {
		nodeCount = int(cluster.MetaData().NodeCount)
	}
	if nodeCount < 1 {
		return false, fmt.Errorf("node-count parameter must be number greater than 0; preferrable 2 or more")
	}
	clusterRequest := bkr.scheduler.NewRequest(cluster, int(nodeCount))
	bkr.scheduler.Execute(clusterRequest)
	return false, nil
}
