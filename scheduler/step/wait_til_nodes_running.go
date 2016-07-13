package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/patronidata"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// WaitTilNodesRunning blocks until expected number of nodes are available and running
type WaitTilNodesRunning struct {
	clusterModel *state.ClusterStateModel
	patroni      *patronidata.Patroni
	logger       lager.Logger
}

// NewWaitTilNodesRunning creates a WaitTilNodesRunning command
func NewWaitTilNodesRunning(clusterModel *state.ClusterStateModel, patroni *patronidata.Patroni, logger lager.Logger) Step {
	return WaitTilNodesRunning{
		clusterModel: clusterModel,
		patroni:      patroni,
		logger:       logger,
	}
}

// StepType prints the type of step
func (step WaitTilNodesRunning) StepType() string {
	return "WaitTilNodesRunning"
}

// Perform runs the Step action upon the Cluster
func (step WaitTilNodesRunning) Perform() (err error) {
	logger := step.logger
	logger.Info("wait-til-nodes-running.perform", lager.Data{"instance-id": step.clusterModel.InstanceID()})

	step.clusterModel.PlanStepStarted("Waiting for all nodes to be running")

	instanceID := step.clusterModel.InstanceID()
	nodesCount := step.clusterModel.NodeCount()

	err = step.patroni.WaitTilClusterMembersRunning(instanceID, nodesCount)
	if err != nil {
		logger.Error("wait-til-nodes-running.perform.error", err, lager.Data{"instance-id": step.clusterModel.InstanceID()})
		return err
	}

	return nil
}
