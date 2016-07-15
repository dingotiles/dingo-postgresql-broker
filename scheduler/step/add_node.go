package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/cells"
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	clusterModel      *state.ClusterModel
	patroni           *patroni.Patroni
	availableBackends backend.Backends
	logger            lager.Logger
	cellsHealth       cells.Cells
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(clusterModel *state.ClusterModel, patroni *patroni.Patroni,
	availableBackends backend.Backends, logger lager.Logger) Step {
	return AddNode{
		clusterModel:      clusterModel,
		patroni:           patroni,
		availableBackends: availableBackends,
		logger:            logger,
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

	// 4. Send requests to sortedBackends until one says OK; else fail
	var provisionedNode structs.Node
	for _, cell := range cellsToTry {
		provisionedNode, err = cell.ProvisionNode(clusterStateData, step.logger)
		logBackend := lager.Data{
			"uri":  cell.URI,
			"guid": cell.ID,
			"az":   cell.AvailabilityZone,
		}
		if err == nil {
			logger.Info("add-node.perform.sorted-backends.selected", logBackend)
			break
		} else {
			logger.Error("add-node.perform.sorted-backends.skipped", err, logBackend)
		}
	}
	if err != nil {
		// no sortedBackends available to run a cluster
		logger.Error("add-node.perform.sorted-backends.unavailable", err, lager.Data{"summary": "no backends available to run a container"})
		return err
	}
	// 5. Store node in KV states/<cluster>/nodes/<node>/backend -> backend uuid
	err = step.clusterModel.AddNode(provisionedNode)
	if err != nil {
		logger.Error("add-node.perform.add-node", err, lager.Data{"summary": "no sorted-backends available to run a cluster"})
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
