package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/patronidata"
)

type Scheduler interface {
	RunCluster(structs.ClusterState, patronidata.ClusterDataWrapper, structs.ClusterFeatures) (structs.ClusterState, error)
	StopCluster(structs.ClusterState, patronidata.ClusterDataWrapper) (structs.ClusterState, error)
	VerifyClusterFeatures(structs.ClusterFeatures) error
}

type State interface {
	ClusterExists(structs.ClusterID) bool
	SaveCluster(structs.ClusterState) error
	LoadCluster(structs.ClusterID) (structs.ClusterState, error)
	DeleteCluster(structs.ClusterID) error
}

type Router interface {
	AllocatePort() (int, error)
	AssignPortToCluster(structs.ClusterID, int) error
	RemoveClusterAssignment(structs.ClusterID) error
}
