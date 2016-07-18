package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/pivotal-golang/lager"
)

// WaitForLeader blocks until leader elected and active
type WaitForLeader struct {
	clusterModel interfaces.ClusterModel
	patroni      *patroni.Patroni
	logger       lager.Logger
}

// NewWaitForLeader creates a WaitForLeader command
func NewWaitForLeader(clusterModel interfaces.ClusterModel, patroni *patroni.Patroni, logger lager.Logger) Step {
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

	instanceID := step.clusterModel.InstanceID()

	err = step.patroni.WaitForLeader(instanceID)
	if err != nil {
		logger.Error("wait-for-leader.perform.error", err, lager.Data{"instance-id": step.clusterModel.InstanceID()})
		return err
	}

	// TODO: Annoyingly, just because patroni agent thinks it is the leader role, doesn't mean
	// that the postgresql is ready yet as a read/write leader.
	// Perhaps need to use PostgreSQL to poll/test writability? Some other method?

	return nil
}
