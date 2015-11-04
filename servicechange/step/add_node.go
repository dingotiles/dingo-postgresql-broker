package step

// AddNode instructs a new cluster node be added
type AddNode struct {
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(nodeSize int) Step {
	return AddNode{}
}
