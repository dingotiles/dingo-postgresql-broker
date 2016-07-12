package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/patronidata"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// LastOperation returns the status of the last operation on a service instance
// TODO: currently assumes all nodes created; if only one at a time then "success" might be too early
// Perhaps AddNode, AddNode, WaitForAllNodesRunning
// Also, LastOperation may need more information about what its waiting for. What if [Add, Add, Remove, Remove]?
// Use /state and /state/errored
func (bkr *Broker) LastOperation(instanceID string) (resp brokerapi.LastOperationResponse, err error) {
	return bkr.lastOperation(structs.ClusterID(instanceID))
}

func (bkr *Broker) lastOperation(instanceID structs.ClusterID) (resp brokerapi.LastOperationResponse, err error) {
	logger := bkr.newLoggingSession("last-opration", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	cluster, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		logger.Error("load-cluster.error", err)
		return brokerapi.LastOperationResponse{State: brokerapi.LastOperationFailed, Description: err.Error()}, err
	}
	if cluster.ErrorMsg != "" {
		return brokerapi.LastOperationResponse{State: brokerapi.LastOperationFailed, Description: cluster.ErrorMsg}, nil
	}

	patroni, _ := patronidata.NewPatroni(bkr.etcdConfig, logger)
	clusterStatus, allRunning, err := patroni.ClusterMembersRunningStates(structs.ClusterID(instanceID), cluster.NodeCount())

	state := brokerapi.LastOperationInProgress
	if allRunning {
		state = brokerapi.LastOperationSucceeded
	}

	return brokerapi.LastOperationResponse{State: state, Description: clusterStatus}, err
}
