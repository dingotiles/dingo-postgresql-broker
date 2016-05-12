package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

type InitCluster struct {
	cluster *state.Cluster
	logger  lager.Logger
}

func NewStepInitCluster(cluster *state.Cluster, logger lager.Logger) Step {
	return InitCluster{cluster: cluster, logger: logger}
}

// StepType prints the type of step
func (step InitCluster) StepType() string {
	return "InitCluster"
}

func (step InitCluster) Perform(backends []*config.Backend) (err error) {
	step.logger.Info("init-cluster.perform", lager.Data{"instance-id": step.cluster.MetaData().InstanceID, "plan-id": step.cluster.MetaData().PlanID})

	err = step.cluster.Init()
	return
}
