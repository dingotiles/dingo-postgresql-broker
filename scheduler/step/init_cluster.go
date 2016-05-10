package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/pivotal-golang/lager"
)

type InitCluster struct {
	cluster *cluster.Cluster
}

func NewStepInitCluster(cluster *cluster.Cluster) Step {
	return InitCluster{cluster: cluster}
}

// StepType prints the type of step
func (step InitCluster) StepType() string {
	return "InitCluster"
}

func (step InitCluster) Perform() (err error) {
	logger := step.cluster.Logger
	logger.Info("init-cluster.perform", lager.Data{"instance-id": step.cluster.Data.InstanceID, "plan-id": step.cluster.Data.PlanID})

	err = step.cluster.Init()
	return
}
