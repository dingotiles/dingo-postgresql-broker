package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
)

// ClusterModel provides a clean .Save() wrapper around a ClusterState for a given State backend
type ClusterModel struct {
	state   interfaces.State
	cluster structs.ClusterState
}

func NewClusterModel(state interfaces.State, cluster structs.ClusterState) *ClusterModel {
	return &ClusterModel{
		state:   state,
		cluster: cluster,
	}
}

func (m *ClusterModel) save() error {
	return m.state.SaveCluster(m.cluster)
}

// SchedulingError stores the failure message for a scheduled Plan
// This will be shown to end users via /last_operation endpoint
func (m *ClusterModel) SchedulingError(err error) error {
	m.cluster.SchedulingInfo.LastMessage = err.Error()
	m.cluster.SchedulingInfo.Status = structs.SchedulingStatusFailed
	return m.save()
}

// SchedulingMessage stores an arbitrary status message
// This will be shown to end users via /last_operation endpoint
func (m *ClusterModel) SchedulingMessage(msg string) error {
	m.cluster.SchedulingInfo.LastMessage = msg
	m.cluster.SchedulingInfo.Status = structs.SchedulingStatusInProgress
	return m.save()
}

func (m *ClusterModel) BeginScheduling(steps int) error {
	m.cluster.SchedulingInfo.Steps = steps
	m.cluster.SchedulingInfo.CompletedSteps = 0
	m.cluster.SchedulingInfo.LastMessage = "In Progress..."
	m.cluster.SchedulingInfo.Status = structs.SchedulingStatusInProgress
	return m.save()
}

func (m *ClusterModel) SchedulingStepCompleted() error {
	m.cluster.SchedulingInfo.CompletedSteps += 1
	if m.cluster.SchedulingInfo.CompletedSteps == m.cluster.SchedulingInfo.Steps {
		m.cluster.SchedulingInfo.Status = structs.SchedulingStatusSuccess
		m.cluster.SchedulingInfo.LastMessage = "Scheduling Completed"
	}

	return m.save()
}

func (m *ClusterModel) SchedulingStepStarted(stepType string) error {
	m.cluster.SchedulingInfo.LastMessage = fmt.Sprintf("Perfoming Step (%d/%d): %s",
		m.cluster.SchedulingInfo.CompletedSteps+1,
		m.cluster.SchedulingInfo.Steps,
		stepType)
	return m.save()
}

func (m *ClusterModel) SchedulingInfo() structs.SchedulingInfo {
	return m.cluster.SchedulingInfo
}

func (m *ClusterModel) ClusterState() structs.ClusterState {
	return m.cluster
}

func (m *ClusterModel) InstanceID() structs.ClusterID {
	return m.cluster.InstanceID
}

func (m *ClusterModel) AllocatedPort() int {
	return m.cluster.AllocatedPort
}

func (m *ClusterModel) NodeCount() int {
	return m.cluster.NodeCount()
}

func (m *ClusterModel) Nodes() []*structs.Node {
	return m.cluster.Nodes
}

func (m *ClusterModel) AddNode(node structs.Node) error {
	m.cluster.AddNode(node)
	return m.save()
}

func (m *ClusterModel) RemoveNode(node *structs.Node) error {
	m.cluster.RemoveNode(node)
	return m.save()
}

func (m *ClusterModel) UpdateCredentials(creds *structs.ClusterRecreationData) error {
	m.cluster.AdminCredentials = creds.AdminCredentials
	m.cluster.SuperuserCredentials = creds.SuperuserCredentials
	m.cluster.AppCredentials = creds.AppCredentials
	return m.save()
}
