package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Update service instance
func (bkr *Broker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, acceptsIncomplete bool) (async bool, err error) {
	logger := bkr.newLoggingSession("update", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	if err := bkr.assertUpdatePrecondition(instanceID); err != nil {
		logger.Error("preconditions.error", err)
		return false, err
	}

	cluster, err := bkr.state.LoadCluster(instanceID)

	var nodeCount int
	if updateDetails.Parameters["node-count"] != nil {
		rawNodeCount := updateDetails.Parameters["node-count"]
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

func (bkr *Broker) assertUpdatePrecondition(instanceID string) error {
	if bkr.state.ClusterExists(instanceID) == false {
		return fmt.Errorf("Service instance %s doesn't exist", instanceID)
	}
	return nil
}
