package step

import (
	"fmt"
	"math/rand"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/pivotal-golang/lager"
)

// RemoveNode instructs cluster to delete a node, starting with replicas
type RemoveNode struct {
	cluster  structs.Cluster
	backends backend.Backends
	logger   lager.Logger
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode(cluster structs.Cluster, backends backend.Backends, logger lager.Logger) Step {
	return RemoveNode{cluster: cluster, backends: backends, logger: logger}
}

// StepType prints the type of step
func (step RemoveNode) StepType() string {
	return "RemoveNode"
}

// Perform runs the Step action to modify the Cluster
func (step RemoveNode) Perform() (err error) {
	logger := step.logger

	// 1. Get list of replicas and pick a random one; else pick a random master
	nodes := step.cluster.Nodes()
	nodeToRemove := randomReplicaNode(nodes)

	backend := step.backends.Get(nodeToRemove.BackendID)
	if backend == nil {
		err = fmt.Errorf("Internal error: node assigned to a backend that no longer exists")
		logger.Error("remove-node.perform", err)
		return
	}

	logger.Info("remove-node.perform", lager.Data{
		"instance-id": step.cluster.MetaData().InstanceID,
		"node-uuid":   nodeToRemove.ID,
		"backend":     backend.ID,
	})

	err = backend.DeprovisionNode(nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.cluster.RemoveNode(nodeToRemove)
	if err != nil {
		logger.Error("remove-node.nodes-delete", err)
	}
	return
}

// currently random any node, doesn't have to be a replica
func randomReplicaNode(nodes []*structs.Node) *structs.Node {
	n := rand.Intn(len(nodes))
	return nodes[n]
}
