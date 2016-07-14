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
func (model *ClusterModel) SchedulingError(err error) error {
	model.cluster.SchedulingInfo.LastMessage = err.Error()
	model.cluster.SchedulingInfo.Error = true
	model.cluster.SchedulingInfo.Status = structs.SchedulingStatusFailed
	return model.save()
}

func (model *ClusterModel) BeginScheduling(steps int) error {
	model.cluster.SchedulingInfo.Steps = steps
	model.cluster.SchedulingInfo.CompletedSteps = 0
	model.cluster.SchedulingInfo.LastMessage = "In Progress..."
	model.cluster.SchedulingInfo.Error = false
	model.cluster.SchedulingInfo.Status = structs.SchedulingStatusInProgress
	return model.save()
}

func (model *ClusterModel) SchedulingStepCompleted() error {
	model.cluster.SchedulingInfo.CompletedSteps += 1
	if model.cluster.SchedulingInfo.CompletedSteps == model.cluster.SchedulingInfo.Steps {
		model.cluster.SchedulingInfo.Status = structs.SchedulingStatusSuccess
		model.cluster.SchedulingInfo.LastMessage = "Scheduling Completed"
	}

	return model.save()
}

func (model *ClusterModel) SchedulingStepStarted(stepType string) error {
	model.cluster.SchedulingInfo.LastMessage = fmt.Sprintf("Perfoming Step: '%s'", stepType)
	return model.save()
}

func (model *ClusterModel) SchedulingInfo() structs.SchedulingInfo {
	return model.cluster.SchedulingInfo
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
