package step

// RemoveNode instructs cluster to delete a node, starting with replicas
type RemoveNode struct {
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode() Step {
	return RemoveNode{}
}

// Perform runs the Step action to modify the Cluster
func (step RemoveNode) Perform() error {
	// logger.Info("add-step.perform", lager.Data{"implemented": false, "step": fmt.Sprintf("%#v", step)})
	return nil
}
