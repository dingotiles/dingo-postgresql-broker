package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// LastOperation returns the status of the last operation on a service instance
// This should not currently be called as Provision() blocks until cluster is running
// CLEANUP: can remove code in future.
func (bkr *Broker) LastOperation(instanceID string) (resp brokerapi.LastOperationResponse, err error) {
	cluster := serviceinstance.NewClusterFromProvisionDetails(instanceID, brokerapi.ProvisionDetails{}, bkr.etcdClient, bkr.config, bkr.logger)
	err = cluster.Load()
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
