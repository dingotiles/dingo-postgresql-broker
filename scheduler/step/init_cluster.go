package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/pivotal-golang/lager"
)

type InitCluster struct {
	cluster *cluster.Cluster
	logger  lager.Logger
}

func NewStepInitCluster(cluster *cluster.Cluster, logger lager.Logger) Step {
	return InitCluster{cluster: cluster, logger: logger}
}

// StepType prints the type of step
func (step InitCluster) StepType() string {
	return "InitCluster"
}

func (step InitCluster) Perform() (err error) {
	step.logger.Info("init-cluster.perform", lager.Data{"instance-id": step.cluster.Data.InstanceID, "plan-id": step.cluster.Data.PlanID})

	err = step.cluster.Init()
	return
}
