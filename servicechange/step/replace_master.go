package step

type ReplaceMaster struct {
	NewNodeSize int
}

func NewStepReplaceMaster(newNodeSize int) Step {
	return ReplaceMaster{NewNodeSize: newNodeSize}
}
