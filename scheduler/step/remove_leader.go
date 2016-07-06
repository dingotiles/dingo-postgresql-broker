package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/pivotal-golang/lager"
)

// RemoveLeader instructs cluster to delete a node, starting with replicas
type RemoveLeader struct {
	nodeToRemove *structs.Node
	cluster      *structs.ClusterState
	backends     backend.Backends
	logger       lager.Logger
}

// NewStepRemoveLeader creates a StepRemoveLeader command
func NewStepRemoveLeader(nodeToRemove *structs.Node, cluster *structs.ClusterState, backends backend.Backends, logger lager.Logger) Step {
	return RemoveLeader{
		nodeToRemove: nodeToRemove,
		cluster:      cluster,
		backends:     backends,
		logger:       logger,
	}
}

// StepType prints the type of step
func (step RemoveLeader) StepType() string {
	return fmt.Sprintf("RemoveLeader(%s)", step.nodeToRemove.ID)
}

// Perform runs the Step action to modify the Cluster
func (step RemoveLeader) Perform() (err error) {
	logger := step.logger

	backend := step.backends.Get(step.nodeToRemove.BackendID)
	if backend == nil {
		err = fmt.Errorf("Internal error: node assigned to a backend that no longer exists")
		logger.Error("remove-node.perform", err)
		return
	}

	logger.Info("remove-node.perform", lager.Data{
		"instance-id": step.cluster.InstanceID,
		"node-uuid":   step.nodeToRemove.ID,
		"backend":     backend.ID,
	})

	err = backend.DeprovisionNode(step.nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.cluster.RemoveNode(step.nodeToRemove)
	if err != nil {
		logger.Error("remove-node.nodes-delete", err)
	}
	return
}
