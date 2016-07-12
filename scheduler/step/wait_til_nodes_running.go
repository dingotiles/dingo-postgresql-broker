package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/patronidata"
	"github.com/pivotal-golang/lager"
)

// WaitTilNodesRunning blocks until expected number of nodes are available and running
type WaitTilNodesRunning struct {
	cluster *structs.ClusterState
	patroni *patronidata.Patroni
	logger  lager.Logger
}

// NewWaitTilNodesRunning creates a WaitTilNodesRunning command
func NewWaitTilNodesRunning(cluster *structs.ClusterState, patroni *patronidata.Patroni, logger lager.Logger) Step {
	return WaitTilNodesRunning{
		cluster: cluster,
		patroni: patroni,
		logger:  logger,
	}
}

// StepType prints the type of step
func (step WaitTilNodesRunning) StepType() string {
	return "WaitTilNodesRunning"
}

// Perform runs the Step action upon the Cluster
func (step WaitTilNodesRunning) Perform() (err error) {
	logger := step.logger
	logger.Info("wait-til-nodes-running.perform", lager.Data{"instance-id": step.cluster.InstanceID})

	nodes := step.cluster.Nodes

	err = step.patroni.WaitTilClusterMembersRunning(step.cluster.InstanceID, len(nodes))
	if err != nil {
		logger.Error("wait-til-nodes-running.perform.error", err, lager.Data{"instance-id": step.cluster.InstanceID})
		return err
	}

	return nil
}
