package interfaces

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
)

type Scheduler interface {
	RunCluster(ClusterModel, structs.ClusterFeatures) error
	StopCluster(ClusterModel) error
	VerifyClusterFeatures(structs.ClusterFeatures) error
}

type Router interface {
	AllocatePort() (int, error)
	AssignPortToCluster(structs.ClusterID, int) error
	RemoveClusterAssignment(structs.ClusterID) error
}

type State interface {
	ClusterExists(structs.ClusterID) bool
	SaveCluster(structs.ClusterState) error
	LoadCluster(structs.ClusterID) (structs.ClusterState, error)
	DeleteCluster(structs.ClusterID) error
}

type ClusterModel interface {
	Cluster() structs.ClusterState
	InstanceID() structs.ClusterID
	AllocatedPort() int
	NodeCount() int
	Nodes() []*structs.Node
	AddNode(structs.Node) error
	RemoveNode(*structs.Node) error

	SchedulingError(err error) error
	BeginScheduling(steps int) error
	SchedulingStepCompleted() error
	SchedulingStepStarted(stepType string) error
	SchedulingInfo() structs.SchedulingInfo
}
