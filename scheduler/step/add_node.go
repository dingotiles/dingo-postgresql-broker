package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	clusterModel   *state.ClusterModel
	patroni        *patroni.Patroni
	availableCells cells.Cells
	logger         lager.Logger
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(clusterModel *state.ClusterModel, patroni *patroni.Patroni,
	availableCells cells.Cells, logger lager.Logger) Step {
	return AddNode{
		clusterModel:   clusterModel,
		patroni:        patroni,
		availableCells: availableCells,
		logger:         logger,
	}
}

// StepType prints the type of step
func (step AddNode) StepType() string {
	return "AddNode"
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform() (err error) {
	logger := step.logger
	logger.Info("add-node.perform", lager.Data{"instance-id": step.clusterModel.InstanceID()})

	existingNodes := step.clusterModel.Nodes()
	clusterStateData := step.clusterModel.Cluster()

	cellsToTry, err := step.prioritizeCellsToTry(existingNodes)
	if err != nil {
		logger.Error("add-node.perform.sorted-cells-to-try", err)
		return err
	}
	logger.Info("add-node.perform.sorted-cells-to-try", lager.Data{"cells": cellsToTry})

	// 4. Send requests to sortedCells until one says OK; else fail
	var provisionedNode structs.Node
	for _, cell := range cellsToTry {
		provisionedNode, err = cell.ProvisionNode(clusterStateData, step.logger)
		logCell := lager.Data{
			"uri":  cell.URI,
			"guid": cell.GUID,
			"az":   cell.AvailabilityZone,
		}
		if err == nil {
			logger.Info("add-node.perform.sorted-cells.selected", logCell)
			break
		} else {
			logger.Error("add-node.perform.sorted-cells.skipped", err, logCell)
		}
	}
	if err != nil {
		// no sortedCells available to run a cluster
		logger.Error("add-node.perform.sorted-cells.unavailable", err, lager.Data{"summary": "no cells available to run a container"})
		return err
	}
	err = step.clusterModel.AddNode(provisionedNode)
	if err != nil {
		logger.Error("add-node.perform.add-node", err, lager.Data{"summary": "no sorted-cells available to run a cluster"})
		return err
	}

	// 6. Wait until node registers itself in data store
	logger.Info("add-node.perform.wait-til-exists", lager.Data{"member": provisionedNode.ID})
	err = step.patroni.WaitForMember(step.clusterModel.InstanceID(), provisionedNode.ID)
	if err != nil {
		logger.Error("add-node.perform.wait-til-exists.error", err, lager.Data{"member": provisionedNode.ID})
		return err
	}

	logger.Info("add-node.perform.success", lager.Data{"member": provisionedNode.ID})
	return nil
}
