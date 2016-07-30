package broker

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// LastOperation returns the status of the last operation on a service instance
// TODO: Plan needs to progressively store the state/description/error message; then
// LastOperation fetches it and returns it (rather than doing any calculations of its own)
// TODO: AddNode, AddNode, WaitForAllMembersRunning
func (bkr *Broker) LastOperation(instanceID string) (resp brokerapi.LastOperationResponse, err error) {
	return bkr.lastOperation(structs.ClusterID(instanceID))
}

func (bkr *Broker) lastOperation(instanceID structs.ClusterID) (resp brokerapi.LastOperationResponse, err error) {
	logger := bkr.newLoggingSession("last-opration", lager.Data{"instance-id": instanceID})
	defer logger.Info("done")

	clusterState, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		logger.Error("load-cluster.error", err)
		return brokerapi.LastOperationResponse{State: brokerapi.LastOperationFailed, Description: err.Error()}, err
	}
	return bkr.lastOperationFromSchedulingInfo(clusterState.SchedulingInfo)
}

func (bkr *Broker) lastOperationFromSchedulingInfo(schedulingInfo structs.SchedulingInfo) (resp brokerapi.LastOperationResponse, err error) {
	resp.Description = schedulingInfo.LastMessage

	switch schedulingInfo.Status {
	case structs.SchedulingStatusFailed:
		resp.State = brokerapi.LastOperationFailed
		err = fmt.Errorf(resp.Description)
	case structs.SchedulingStatusSuccess:
		resp.State = brokerapi.LastOperationSucceeded
	case structs.SchedulingStatusInProgress:
		resp.State = brokerapi.LastOperationInProgress
	default:
		resp.State = brokerapi.LastOperationInProgress
		resp.Description = "Preparing..."
	}
	return
}
