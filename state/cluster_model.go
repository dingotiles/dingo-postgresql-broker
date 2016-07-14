package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
)

// ClusterModel provides a clean .Save() wrapper around a ClusterState for a given State backend
type ClusterModel struct {
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

func NewClusterModel(state State, cluster structs.ClusterState) *ClusterModel {
	return &ClusterModel{
		state:   state,
		cluster: cluster,
	}
}

func (model *ClusterModel) save() error {
	return model.state.SaveCluster(model.cluster)
}

// PlanError stores the failure message for a scheduled Plan
// This will be shown to end users via /last_operation endpoint
func (model *ClusterModel) PlanError(err error) error {
	model.cluster.ErrorMsg = err.Error()
	return model.save()
}

func (model *ClusterModel) NewClusterPlan(steps int) error {
	model.cluster.Plan.Steps = steps
	model.cluster.Plan.CompletedSteps = 0
	model.cluster.Plan.Message = ""
	return model.save()
}

func (model *ClusterModel) PlanStepComplete() error {
	model.cluster.Plan.CompletedSteps += 1
	return model.save()
}

func (model *ClusterModel) PlanStepStarted(msg string) error {
	model.cluster.Plan.Message = msg
	return model.save()
}

func (model *ClusterModel) CurrentPlanStatus() (status *PlanStatus) {
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

func (model *ClusterModel) Cluster() structs.ClusterState {
	return model.cluster
}

func (model *ClusterModel) InstanceID() structs.ClusterID {
	return model.cluster.InstanceID
}

func (model *ClusterModel) AllocatedPort() int {
	return model.cluster.AllocatedPort
}

func (model *ClusterModel) NodeCount() int {
	return model.cluster.NodeCount()
}

func (model *ClusterModel) Nodes() []*structs.Node {
	return model.cluster.Nodes
}

func (model *ClusterModel) AddNode(node structs.Node) error {
	model.cluster.AddNode(node)
	return model.save()
}

func (model *ClusterModel) RemoveNode(node *structs.Node) error {
	model.cluster.RemoveNode(node)
	return model.save()
}
