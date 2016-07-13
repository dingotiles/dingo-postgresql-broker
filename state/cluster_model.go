package state

import "github.com/dingotiles/dingo-postgresql-broker/broker/structs"

// ClusterStateModel provides a clean .Save() wrapper around a ClusterState for a given State backend
type ClusterStateModel struct {
	state   State
	cluster structs.ClusterState
}

func NewClusterStateModel(state State, cluster structs.ClusterState) *ClusterStateModel {
	return &ClusterStateModel{
		state:   state,
		cluster: cluster,
	}
}

func (model *ClusterStateModel) Save() error {
	return model.state.SaveCluster(model.cluster)
}

// ResetClusterPlan refreshes previous Plan state/error messages
// in preparation for commencing new Plan
func (model *ClusterStateModel) ResetClusterPlan() error {
	model.cluster.ErrorMsg = ""
	return model.Save()
}

// PlanError stores the failure message for a scheduled Plan
// This will be shown to end users via /last_operation endpoint
func (model *ClusterStateModel) PlanError(err error) error {
	model.cluster.ErrorMsg = err.Error()
	return model.Save()
}

func (model *ClusterStateModel) Cluster() *structs.ClusterState {
	return &model.cluster
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
	return model.cluster.AddNode(node)
}

func (model *ClusterStateModel) RemoveNode(node *structs.Node) error {
	return model.cluster.RemoveNode(node)
}
