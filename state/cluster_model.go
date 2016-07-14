package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
)

// ClusterStateModel provides a clean .Save() wrapper around a ClusterState for a given State backend
type ClusterStateModel struct {
	state   State
	cluster structs.ClusterState
}

type PlanStatus struct {
	Status  string
	Message string
}

const (
	PlanStatusUnknown    = ""
	PlanStatusSuccess    = "success"
	PlanStatusInProgress = "in-progress"
	PlanStatusFailed     = "failed"
)

func NewClusterStateModel(state State, cluster structs.ClusterState) *ClusterStateModel {
	return &ClusterStateModel{
		state:   state,
		cluster: cluster,
	}
}

func (model *ClusterStateModel) save() error {
	return model.state.SaveCluster(model.cluster)
}

// PlanError stores the failure message for a scheduled Plan
// This will be shown to end users via /last_operation endpoint
func (model *ClusterStateModel) PlanError(err error) error {
	model.cluster.ErrorMsg = err.Error()
	return model.save()
}

func (model *ClusterStateModel) NewClusterPlan(steps int) error {
	model.cluster.Plan.Steps = steps
	model.cluster.Plan.CompletedSteps = 0
	model.cluster.Plan.Message = ""
	return model.save()
}

func (model *ClusterStateModel) PlanStepComplete() error {
	model.cluster.Plan.CompletedSteps += 1
	return model.save()
}

func (model *ClusterStateModel) PlanStepStarted(msg string) error {
	model.cluster.Plan.Message = msg
	return model.save()
}

func (model *ClusterStateModel) CurrentPlanStatus() (status *PlanStatus) {
	msg := fmt.Sprintf("%s %d/%d", model.cluster.Plan.Message, model.cluster.Plan.CompletedSteps, model.cluster.Plan.Steps)
	status = &PlanStatus{
		Status:  PlanStatusInProgress,
		Message: msg,
	}
	if model.cluster.ErrorMsg != "" {
		status.Message = fmt.Sprintf("Error: %s %d/%d", model.cluster.ErrorMsg, model.cluster.Plan.Steps, model.cluster.Plan.Steps)
		status.Status = PlanStatusFailed
		return
	}
	if model.cluster.Plan.Steps == 0 {
		status.Status = PlanStatusUnknown
		return
	}
	if model.cluster.Plan.CompletedSteps == model.cluster.Plan.Steps {
		status.Message = fmt.Sprintf("Completed %d/%d", model.cluster.Plan.CompletedSteps, model.cluster.Plan.Steps)
		status.Status = PlanStatusSuccess
		return
	}

	return
}

func (model *ClusterStateModel) Cluster() structs.ClusterState {
	return model.cluster
}

func (model *ClusterStateModel) InstanceID() structs.ClusterID {
	return model.cluster.InstanceID
}

func (model *ClusterStateModel) AllocatedPort() int {
	return model.cluster.AllocatedPort
}

func (model *ClusterStateModel) NodeCount() int {
	return model.cluster.NodeCount()
}

func (model *ClusterStateModel) Nodes() []*structs.Node {
	return model.cluster.Nodes
}

func (model *ClusterStateModel) AddNode(node structs.Node) error {
	model.cluster.AddNode(node)
	return model.save()
}

func (model *ClusterStateModel) RemoveNode(node *structs.Node) error {
	model.cluster.RemoveNode(node)
	return model.save()
}
