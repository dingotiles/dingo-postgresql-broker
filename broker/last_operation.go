package broker

import (
	"fmt"

	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// LastOperation returns the status of the last operation on a service instance
// brokerapi:
// const LastOperationInProgress = "in progress"
// const LastOperationFailed = "failed"
// const LastOperationSucceeded = "succeeded"
func (bkr *Broker) LastOperation(instanceID string) (resp brokerapi.LastOperationResponse, err error) {
	cluster := serviceinstance.NewCluster(instanceID, brokerapi.ProvisionDetails{}, bkr.EtcdClient, bkr.Config, bkr.Logger)
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
