package step

import (
	"fmt"

	"github.com/pivotal-golang/lager"
)

type ReplaceMaster struct {
	NewNodeSize int
}

func NewStepReplaceMaster(newNodeSize int) Step {
	return ReplaceMaster{NewNodeSize: newNodeSize}
}

// Perform runs the Step action to modify the Cluster
func (step ReplaceMaster) Perform(logger lager.Logger) {
	logger.Info("add-step.perform", lager.Data{"implemented": false, "step": fmt.Sprintf("%#v", step)})

}
