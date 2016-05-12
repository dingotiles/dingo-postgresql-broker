package step

import "github.com/dingotiles/dingo-postgresql-broker/config"

// ReplaceReplica describes an action to create a new resized replica node, then destroy an older one
type ReplaceReplica struct {
	CurrentNodeSize int
	NewNodeSize     int
}

// NewStepReplaceReplica describes an action to create a new resized replica node, then destroy an older one
func NewStepReplaceReplica(currentNodeSize int, newNodeSize int) Step {
	return ReplaceReplica{CurrentNodeSize: currentNodeSize, NewNodeSize: newNodeSize}
}

// StepType prints the type of step
func (step ReplaceReplica) StepType() string {
	return "ReplaceReplica"
}

// Perform runs the Step action to modify the Cluster
func (step ReplaceReplica) Perform(backends []*config.Backend) error {
	// logger.Info("add-step.perform", lager.Data{"implemented": false, "step": fmt.Sprintf("%#v", step)})
	return nil
}
