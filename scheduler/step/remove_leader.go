package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// RemoveLeader instructs cluster to delete a node, starting with replicas
type RemoveLeader struct {
	nodeToRemove *structs.Node
	clusterModel *state.ClusterModel
	cells        cells.Cells
	logger       lager.Logger
}

// NewStepRemoveLeader creates a StepRemoveLeader command
func NewStepRemoveLeader(nodeToRemove *structs.Node, clusterModel *state.ClusterModel, cells cells.Cells, logger lager.Logger) Step {
	return RemoveLeader{
		nodeToRemove: nodeToRemove,
		clusterModel: clusterModel,
		cells:        cells,
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

	cell := step.cells.Get(step.nodeToRemove.CellGUID)
	if cell == nil {
		err = fmt.Errorf("Internal error: node assigned to a cell that no longer exists (%s)", step.nodeToRemove.CellGUID)
		logger.Error("remove-leader.perform", err)
		return
	}

	logger.Info("remove-leader.perform", lager.Data{
		"instance-id": step.clusterModel.InstanceID(),
		"node-uuid":   step.nodeToRemove.ID,
		"cell":        cell.GUID,
	})

	err = cell.DeprovisionNode(step.nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.clusterModel.RemoveNode(step.nodeToRemove)
	if err != nil {
		logger.Error("remove-leader.nodes-delete", err)
	}

	return
}
