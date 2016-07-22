package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/pivotal-golang/lager"
)

// WaitForAllMembers blocks until expected number of nodes are available and running
type WaitForAllMembers struct {
	clusterModel interfaces.ClusterModel
	patroni      interfaces.Patroni
	logger       lager.Logger
}

// NewWaitForAllMembers creates a WaitForAllMembers command
func NewWaitForAllMembers(clusterModel interfaces.ClusterModel, patroni interfaces.Patroni, logger lager.Logger) Step {
	return WaitForAllMembers{
		clusterModel: clusterModel,
		patroni:      patroni,
		logger:       logger,
	}
}

// StepType prints the type of step
func (step WaitForAllMembers) StepType() string {
	return "WaitForAllMembers"
}

// Perform runs the Step action upon the Cluster
func (step WaitForAllMembers) Perform() (err error) {
	logger := step.logger
	logger.Info("wait-til-nodes-running.perform", lager.Data{"instance-id": step.clusterModel.InstanceID()})

	instanceID := step.clusterModel.InstanceID()
	nodesCount := step.clusterModel.NodeCount()

	err = step.patroni.WaitForAllMembers(instanceID, nodesCount)
	if err != nil {
		logger.Error("wait-til-nodes-running.perform.error", err, lager.Data{"instance-id": step.clusterModel.InstanceID()})
		return err
	}

	return nil
}
