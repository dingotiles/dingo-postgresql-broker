package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/patronidata"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// WaitForLeader blocks until leader elected and active
type WaitForLeader struct {
	clusterModel *state.ClusterStateModel
	patroni      *patronidata.Patroni
	logger       lager.Logger
}

// NewWaitForLeader creates a WaitForLeader command
func NewWaitForLeader(clusterModel *state.ClusterStateModel, patroni *patronidata.Patroni, logger lager.Logger) Step {
	return WaitForLeader{
		clusterModel: clusterModel,
		patroni:      patroni,
		logger:       logger,
	}
}

// StepType prints the type of step
func (step WaitForLeader) StepType() string {
	return "WaitForLeader"
}

// Perform runs the Step action upon the Cluster
func (step WaitForLeader) Perform() (err error) {
	logger := step.logger
	logger.Info("wait-for-leader.perform", lager.Data{"instance-id": step.clusterModel.InstanceID()})

	step.clusterModel.PlanStepStarted("Waiting for leader")

	instanceID := step.clusterModel.InstanceID()

	err = step.patroni.WaitForLeader(instanceID)
	if err != nil {
		logger.Error("wait-for-leader.perform.error", err, lager.Data{"instance-id": step.clusterModel.InstanceID()})
		return err
	}

	return nil
}
