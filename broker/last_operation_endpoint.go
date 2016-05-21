package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// LastOperation returns the status of the last operation on a service instance
// This should not currently be called as Provision() blocks until cluster is running
// CLEANUP: can remove code in future.
func (bkr *Broker) LastOperation(instanceID string) (resp brokerapi.LastOperationResponse, err error) {
	logger := bkr.newLoggingSession("last-opration", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	clusterStatus, allRunning, err := patroni.MemberStatus(instanceID, bkr.etcdClient, logger)

	state := brokerapi.LastOperationInProgress
	if allRunning {
		state = brokerapi.LastOperationSucceeded
	}

	return brokerapi.LastOperationResponse{State: state, Description: clusterStatus}, err
}
