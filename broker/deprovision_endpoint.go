package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Deprovision service instance
func (bkr *Broker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, acceptsIncomplete bool) (async bool, err error) {
	logger := bkr.newLoggingSession("deprovision", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	if err = bkr.assertDeprovisionPrecondition(instanceID, details); err != nil {
		logger.Error("preconditions.error", err)
		return false, err
	}

	cluster, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		logger.Error("load-cluster.error", err)
		return false, err
	}

	bkr.scheduler.StopCluster(cluster)
	bkr.state.DeleteCluster(instanceID)
	bkr.router.RemoveClusterAssignment(instanceID)

	return false, nil
}

func (bkr *Broker) assertDeprovisionPrecondition(instanceID string, details brokerapi.DeprovisionDetails) error {
	if bkr.state.ClusterExists(instanceID) == false {
		return fmt.Errorf("Service instance %s doesn't exist", instanceID)
	}

	if details.ServiceID == "" || details.PlanID == "" {
		return fmt.Errorf("API error - provide service_id and plan_id as URL parameters")
	}

	return nil
}
