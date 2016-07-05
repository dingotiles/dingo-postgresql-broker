package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/pivotal-golang/lager"
)

const (
	defaultNodeSize = 20
)

// p.est represents a user-originating p.est to change a service instance (grow, scale, move)
type plan struct {
	clusterState *structs.ClusterState
	newFeatures  structs.ClusterFeatures
	backends     backend.Backends
	newNodeSize  int
	logger       lager.Logger
}

// Newp.est creates a p.est to change a service instance
func (s *Scheduler) newPlan(cluster *structs.ClusterState, features structs.ClusterFeatures) (plan, error) {
	backends, err := s.FilterCellsByGUIDs(features.CellGUIDsForNewNodes)
	if err != nil {
		return plan{}, err
	}
	return plan{
		clusterState: cluster,
		newFeatures:  features,
		backends:     backends,
		logger:       s.logger,
		newNodeSize:  defaultNodeSize,
	}, nil
}

// stepTypes is the ordered sequence of workflow step types to orchestrate a service instance change
func (p plan) stepTypes() []string {
	steps := p.steps()
	stepTypes := make([]string, len(steps))
	for i, step := range steps {
		stepTypes[i] = step.StepType()
	}
	return stepTypes
}

// steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (p plan) steps() []step.Step {
	existingNodeSize := defaultNodeSize
	targetNodeCount := p.newFeatures.NodeCount
	currentNodeCount := p.clusterState.NodeCount()

	steps := []step.Step{}
	if targetNodeCount == 0 {
		for i := currentNodeCount; i > targetNodeCount; i-- {
			steps = append(steps, step.NewStepRemoveNode(p.clusterState, p.backends, p.logger))
		}
	} else if !p.isScalingUp() && !p.isScalingDown() &&
		!p.isScalingIn() && !p.isScalingOut() {
		return steps
	} else if !p.isScalingUp() && !p.isScalingDown() {
		// if only scaling out or in; but not up or down
		if p.isScalingOut() {
			for i := currentNodeCount; i < targetNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(p.clusterState, p.backends, p.logger))
			}
		}
		if p.isScalingIn() {
			for i := currentNodeCount; i > targetNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(p.clusterState, p.backends, p.logger))
			}
		}
	} else if !p.isScalingIn() && !p.isScalingOut() {
		// if only scaling up or down; but not out or in
		steps = append(steps, step.NewStepReplaceMaster(p.newNodeSize))
		// replace remaining replicas with resized nodes
		for i := 1; i < currentNodeCount; i++ {
			steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, p.newNodeSize))
		}
	} else {
		// changing both horizontal and vertical aspects of cluster
		steps = append(steps, step.NewStepReplaceMaster(p.newNodeSize))
		if p.isScalingOut() {
			for i := 1; i < currentNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, p.newNodeSize))
			}
			for i := currentNodeCount; i < targetNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(p.clusterState, p.backends, p.logger))
			}
		} else if p.isScalingIn() {
			for i := 1; i < targetNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, p.newNodeSize))
			}
			for i := currentNodeCount; i > targetNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(p.clusterState, p.backends, p.logger))
			}
		}
	}
	return steps
}

// isScalingUp is true if smaller nodes
func (p plan) isScalingUp() bool {
	return false
	// return p.newNodeSize != 0 && defaultNodeSize < p.newNodeSize
}

// isScalingDown is true if bigger nodes
func (p plan) isScalingDown() bool {
	return false
	// return p.newNodeSize != 0 && defaultNodeSize > p.newNodeSize
}

// isScalingOut is true if more nodes p.ested
func (p plan) isScalingOut() bool {
	return p.newFeatures.NodeCount != 0 && p.clusterState.NodeCount() < p.newFeatures.NodeCount
}

// isScalingIn is true if fewer nodes p.ested
func (p plan) isScalingIn() bool {
	return p.newFeatures.NodeCount != 0 && p.clusterState.NodeCount() > p.newFeatures.NodeCount
}
