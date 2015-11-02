package servicechange

// Step is a step in a workflow to change a cluster (grow, scale, move)
type Step interface {
}

// StepAddNode instructs a new cluster node be added
type StepAddNode struct {
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(nodeSize int) Step {
	return StepAddNode{}
}

// StepRemoveNode instructs cluster to delete a node, starting with replicas
type StepRemoveNode struct {
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode() Step {
	return StepRemoveNode{}
}

type StepReplaceMaster struct {
	NewNodeSize int
}

func NewStepReplaceMaster(newNodeSize int) Step {
	return StepReplaceMaster{NewNodeSize: newNodeSize}
}

type StepReplaceReplica struct {
	CurrentNodeSize int
	NewNodeSize     int
}

func NewStepReplaceReplica(currentNodeSize int, newNodeSize int) Step {
	return StepReplaceReplica{CurrentNodeSize: currentNodeSize, NewNodeSize: newNodeSize}
}
