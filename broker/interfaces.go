package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/state"
)

type Scheduler interface {
	RunCluster(*state.ClusterStateModel, structs.ClusterFeatures) error
	StopCluster(*state.ClusterStateModel) error
	VerifyClusterFeatures(structs.ClusterFeatures) error
}

type Router interface {
	AllocatePort() (int, error)
	AssignPortToCluster(structs.ClusterID, int) error
	RemoveClusterAssignment(structs.ClusterID) error
}
