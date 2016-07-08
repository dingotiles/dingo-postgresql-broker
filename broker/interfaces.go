package broker

import "github.com/dingotiles/dingo-postgresql-broker/broker/structs"

type Scheduler interface {
	RunCluster(structs.ClusterState, structs.ClusterFeatures) (structs.ClusterState, error)
	StopCluster(structs.ClusterState) (structs.ClusterState, error)
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
