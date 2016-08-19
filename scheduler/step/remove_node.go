package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/pivotal-golang/lager"
)

// RemoveNode instructs cluster to delete a node, starting with replicas
type RemoveNode struct {
	nodeToRemove *structs.Node
	clusterModel interfaces.ClusterModel
	cells        cells.Cells
	logger       lager.Logger
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode(nodeToRemove *structs.Node, clusterModel interfaces.ClusterModel, cells cells.Cells, logger lager.Logger) Step {
	return RemoveNode{
		nodeToRemove: nodeToRemove,
		clusterModel: clusterModel,
		cells:        cells,
		logger:       logger,
	}
}

// StepType prints the type of step
func (step RemoveNode) StepType() string {
	return fmt.Sprintf("RemoveNode(%s)", step.nodeToRemove.ID)
}

// Perform runs the Step action to modify the Cluster
func (step RemoveNode) Perform() (err error) {
	logger := step.logger

	cell := step.cells.Get(step.nodeToRemove.CellGUID)
	if cell == nil {
		err = fmt.Errorf("Internal error: node assigned to a cell that no longer exists (%s)", step.nodeToRemove.CellGUID)
		logger.Error("remove-node.perform", err)
		return
	}

	logger.Info("remove-node.perform", lager.Data{
		"instance-id": step.clusterModel.InstanceID(),
		"node-uuid":   step.nodeToRemove.ID,
		"cell":        cell.GUID,
	})

	err = cell.DeprovisionNode(step.clusterModel.ClusterState(), step.nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.clusterModel.RemoveNode(step.nodeToRemove)
	if err != nil {
		logger.Error("remove-node.nodes-delete", err)
	}
	return
}
