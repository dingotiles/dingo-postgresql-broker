package broker

import "github.com/dingotiles/dingo-postgresql-broker/broker/structs"

type Scheduler interface {
	RunCluster(structs.ClusterState, structs.ClusterFeatures) (structs.ClusterState, error)
	StopCluster(structs.ClusterState) (structs.ClusterState, error)
	VerifyClusterFeatures(structs.ClusterFeatures) error
}

type State interface {
	ClusterExists(instanceID string) bool
	SaveCluster(cluster structs.ClusterState) error
	LoadCluster(instanceID string) (structs.ClusterState, error)
	DeleteCluster(instanceID string) error
}

type Router interface {
	AllocatePort() (int, error)
	AssignPortToCluster(clusterID string, port int) error
	RemoveClusterAssignment(clusterID string) error
}
