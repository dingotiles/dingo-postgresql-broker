package step

import (
	"fmt"

	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(nodeSize int) Step {
	return AddNode{}
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform(logger lager.Logger) {
	logger.Info("add-step.perform", lager.Data{"implemented": true, "step": fmt.Sprintf("%#v", step)})
}
