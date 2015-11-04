package step

type ReplaceReplica struct {
	CurrentNodeSize int
	NewNodeSize     int
}

func NewStepReplaceReplica(currentNodeSize int, newNodeSize int) Step {
	return ReplaceReplica{CurrentNodeSize: currentNodeSize, NewNodeSize: newNodeSize}
}
