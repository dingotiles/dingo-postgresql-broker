package step

// RemoveNode instructs cluster to delete a node, starting with replicas
type RemoveNode struct {
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode() Step {
	return RemoveNode{}
}
