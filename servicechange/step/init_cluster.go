package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/serviceinstance"
	"github.com/pivotal-golang/lager"
)

type InitCluster struct {
	cluster *serviceinstance.Cluster
}

func NewStepInitCluster(cluster *serviceinstance.Cluster) Step {
	return InitCluster{cluster: cluster}
}

// StepType prints the type of step
func (step InitCluster) StepType() string {
	return "InitCluster"
}

func (step InitCluster) Perform() (err error) {
	logger := step.cluster.Logger
	logger.Info("init-cluster.perform", lager.Data{"instance-id": step.cluster.Data.InstanceID, "plan-id": step.cluster.Data.PlanID})

	key := fmt.Sprintf("/serviceinstances/%s/plan_id", step.cluster.Data.InstanceID)
	_, err = step.cluster.EtcdClient.Set(key, step.cluster.Data.PlanID, 0)

	return
}
