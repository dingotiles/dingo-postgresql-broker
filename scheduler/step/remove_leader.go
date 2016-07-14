package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// RemoveLeader instructs cluster to delete a node, starting with replicas
type RemoveLeader struct {
	nodeToRemove *structs.Node
	clusterModel *state.ClusterModel
	backends     backend.Backends
	logger       lager.Logger
}

// NewStepRemoveLeader creates a StepRemoveLeader command
func NewStepRemoveLeader(nodeToRemove *structs.Node, clusterModel *state.ClusterModel, backends backend.Backends, logger lager.Logger) Step {
	return RemoveLeader{
		nodeToRemove: nodeToRemove,
		clusterModel: clusterModel,
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

	step.clusterModel.PlanStepStarted("Replacing/removing leader")

	backend := step.backends.Get(step.nodeToRemove.BackendID)
	if backend == nil {
		err = fmt.Errorf("Internal error: node assigned to a backend that no longer exists (%s)", step.nodeToRemove.BackendID)
		logger.Error("remove-leader.perform", err)
		return
	}

	logger.Info("remove-leader.perform", lager.Data{
		"instance-id": step.clusterModel.InstanceID(),
		"node-uuid":   step.nodeToRemove.ID,
		"backend":     backend.ID,
	})

	err = backend.DeprovisionNode(step.nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.clusterModel.RemoveNode(step.nodeToRemove)
	if err != nil {
		logger.Error("remove-leader.nodes-delete", err)
	}

	return
}
