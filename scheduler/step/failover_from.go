package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/pivotal-golang/lager"
)

// WaitForLeader blocks until leader elected and active
type FailoverFrom struct {
	leaderID     string
	clusterModel interfaces.ClusterModel
	patroni      interfaces.Patroni
	logger       lager.Logger
}

// NewStepFailoverFrom creates a FailoverFrom command
func NewStepFailoverFrom(clusterModel interfaces.ClusterModel, leaderID string, patroni interfaces.Patroni, logger lager.Logger) Step {
	return FailoverFrom{
		leaderID:     leaderID,
		clusterModel: clusterModel,
		patroni:      patroni,
		logger:       logger,
	}
}

// StepType prints the type of step
func (step FailoverFrom) StepType() string {
	return fmt.Sprintf("FailoverFrom(%s)", step.leaderID)
}

// Perform runs the Step action upon the Cluster
func (step FailoverFrom) Perform() (err error) {
	logger := step.logger
	logger.Info("failover-from.perform", lager.Data{"instance-id": step.clusterModel.InstanceID(), "leader-id": step.leaderID})

	instanceID := step.clusterModel.InstanceID()

	err = step.patroni.FailoverFrom(instanceID, step.leaderID)
	if err != nil {
		logger.Error("failover-from.perform.error", err, lager.Data{"instance-id": step.clusterModel.InstanceID(), "leader-id": step.leaderID})
		return err
	}

	return nil
}
