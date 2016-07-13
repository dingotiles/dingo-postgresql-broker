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

func (model *ClusterStateModel) Cluster() *structs.ClusterState {
	return &model.cluster
}

func (model *ClusterStateModel) InstanceID() structs.ClusterID {
	return model.cluster.InstanceID
}
