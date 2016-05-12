package step

import "github.com/dingotiles/dingo-postgresql-broker/config"

// ReplaceMaster describes the new master
type ReplaceMaster struct {
	NewNodeSize int
}

// NewStepReplaceMaster prepares to change the service by replacing/upgrading the leader/master
func NewStepReplaceMaster(newNodeSize int) Step {
	return ReplaceMaster{NewNodeSize: newNodeSize}
}

// StepType prints the type of step
func (step ReplaceMaster) StepType() string {
	return "ReplaceMaster"
}

// Perform runs the Step action to modify the Cluster
func (step ReplaceMaster) Perform(backends []*config.Backend) error {
	// logger.Info("add-step.perform", lager.Data{"implemented": false, "step": fmt.Sprintf("%#v", step)})
	return nil

}
