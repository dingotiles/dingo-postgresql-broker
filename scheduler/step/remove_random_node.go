package step

import (
	"fmt"
	"math/rand"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// RemoveRandomNode instructs cluster to delete a node, starting with replicas
type RemoveRandomNode struct {
	clusterModel *state.ClusterStateModel
	backends     backend.Backends
	logger       lager.Logger
}

// NewStepRemoveRandomNode creates a StepRemoveRandomNode command
func NewStepRemoveRandomNode(clusterModel *state.ClusterStateModel, backends backend.Backends, logger lager.Logger) Step {
	return RemoveRandomNode{clusterModel: clusterModel, backends: backends, logger: logger}
}

// StepType prints the type of step
func (step RemoveRandomNode) StepType() string {
	return "RemoveRandomNode"
}

// Perform runs the Step action to modify the Cluster
func (step RemoveRandomNode) Perform() (err error) {
	logger := step.logger

	step.clusterModel.PlanStepStarted("Removing node")

	// 1. Get list of replicas and pick a random one; else pick a random master
	nodes := step.clusterModel.Nodes()
	nodeToRemove := randomReplicaNode(nodes)

	backend := step.backends.Get(nodeToRemove.BackendID)
	if backend == nil {
		err = fmt.Errorf("Internal error: node assigned to a backend that no longer exists (%s)", nodeToRemove.BackendID)
		logger.Error("remove-random-node.perform", err)
		return
	}

	logger.Info("remove-random-node.perform", lager.Data{
		"instance-id": step.clusterModel.InstanceID(),
		"node-uuid":   nodeToRemove.ID,
		"backend":     backend.ID,
	})

	err = backend.DeprovisionNode(nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.clusterModel.RemoveNode(nodeToRemove)
	if err != nil {
		logger.Error("remove-random-node.nodes-delete", err)
	}
	return
}

// currently random any node, doesn't have to be a replica
func randomReplicaNode(nodes []*structs.Node) *structs.Node {
	n := rand.Intn(len(nodes))
	return nodes[n]
}
