package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/state"
)

type Scheduler interface {
	RunCluster(*state.ClusterModel, structs.ClusterFeatures) error
	StopCluster(*state.ClusterModel) error
	VerifyClusterFeatures(structs.ClusterFeatures) error
}

type Router interface {
	AllocatePort() (int, error)
	AssignPortToCluster(structs.ClusterID, int) error
	RemoveClusterAssignment(structs.ClusterID) error
}
