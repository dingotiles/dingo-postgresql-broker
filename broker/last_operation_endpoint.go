package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// LastOperation returns the status of the last operation on a service instance
// This should not currently be called as Provision() blocks until cluster is running
// CLEANUP: can remove code in future.
func (bkr *Broker) LastOperation(instanceID string) (resp brokerapi.LastOperationResponse, err error) {
	logger := bkr.newLoggingSession("last-opration", lager.Data{"instanceID": instanceID})
	defer logger.Info("stop")

	cluster, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		return brokerapi.LastOperationResponse{
			State:       brokerapi.LastOperationFailed,
			Description: fmt.Sprintf("Cannot find service instance %s", instanceID),
		}, err
	}
	clusterStatus, allRunning, err := cluster.MemberStatus()

	state := brokerapi.LastOperationInProgress
	if allRunning {
		state = brokerapi.LastOperationSucceeded
	}

	return brokerapi.LastOperationResponse{State: state, Description: clusterStatus}, err
}