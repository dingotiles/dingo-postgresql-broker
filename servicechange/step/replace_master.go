package step

type ReplaceMaster struct {
	NewNodeSize int
}

func NewStepReplaceMaster(newNodeSize int) Step {
	return ReplaceMaster{NewNodeSize: newNodeSize}
}

func (step ReplaceMaster) StepType() string {
	return "ReplaceMaster"
}

// Perform runs the Step action to modify the Cluster
func (step ReplaceMaster) Perform() error {
	// logger.Info("add-step.perform", lager.Data{"implemented": false, "step": fmt.Sprintf("%#v", step)})
	return nil

}
