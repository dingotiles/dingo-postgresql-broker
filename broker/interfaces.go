package broker

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
)

type Scheduler interface {
	RunCluster(structs.ClusterState, config.Etcd, structs.ClusterFeatures) (structs.ClusterState, error)
	StopCluster(structs.ClusterState, config.Etcd) (structs.ClusterState, error)
	VerifyClusterFeatures(structs.ClusterFeatures) error
}

type Router interface {
	AllocatePort() (int, error)
	AssignPortToCluster(structs.ClusterID, int) error
	RemoveClusterAssignment(structs.ClusterID) error
}
