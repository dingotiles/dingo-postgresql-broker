package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Deprovision service instance
func (bkr *Broker) Deprovision(instanceID string, deprovDetails brokerapi.DeprovisionDetails, acceptsIncomplete bool) (async bool, err error) {
	logger := bkr.newLoggingSession("deprovision", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	if deprovDetails.ServiceID == "" || deprovDetails.PlanID == "" {
		return false, fmt.Errorf("API error - provide service_id and plan_id as URL parameters")
	}
	cluster, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		logger.Error("load-cluster", err)
		return false, err
	}

	cluster.SetTargetNodeCount(0)
	clusterRequest := bkr.scheduler.NewRequest(cluster)
	bkr.scheduler.Execute(clusterRequest)

	bkr.state.DeleteCluster(cluster)

	return false, nil
}
