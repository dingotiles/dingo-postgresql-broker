package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

const (
	defaultNodeSize = 20
)

// Request represents a user-originating request to change a service instance (grow, scale, move)
type Request struct {
	cluster      *state.Cluster
	backends     backend.Backends
	newNodeSize  int
	newNodeCount int
	logger       lager.Logger
}

// NewRequest creates a Request to change a service instance
func (s *Scheduler) NewRequest(cluster *state.Cluster, nodeCount int) Request {
	return Request{
		cluster:      cluster,
		backends:     s.backends,
		newNodeCount: nodeCount,
		newNodeSize:  defaultNodeSize,
		logger:       s.logger,
	}
}

// stepTypes is the ordered sequence of workflow step types to orchestrate a service instance change
func (req Request) stepTypes() []string {
	steps := req.steps()
	stepTypes := make([]string, len(steps))
	for i, step := range steps {
		stepTypes[i] = step.StepType()
	}
	return stepTypes
}

// steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (req Request) steps() []step.Step {
	existingNodeCount := req.cluster.MetaData().TargetNodeCount
	existingNodeSize := defaultNodeSize
	steps := []step.Step{}
	if req.newNodeCount == 0 {
		for i := existingNodeCount; i > req.newNodeCount; i-- {
			steps = append(steps, step.NewStepRemoveNode(req.cluster, req.backends, req.logger))
		}
	} else if !req.isScalingUp() && !req.isScalingDown() &&
		!req.isScalingIn() && !req.isScalingOut() {
		return steps
	} else if req.isInitialProvision() {
		for i := existingNodeCount; i < req.newNodeCount; i++ {
			steps = append(steps, step.NewStepAddNode(req.cluster, req.backends, req.logger))
		}
	} else if !req.isScalingUp() && !req.isScalingDown() {
		// if only scaling out or in; but not up or down
		if req.isScalingOut() {
			for i := existingNodeCount; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.backends, req.logger))
			}
		}
		if req.isScalingIn() {
			for i := existingNodeCount; i > req.newNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.backends, req.logger))
			}
		}
	} else if !req.isScalingIn() && !req.isScalingOut() {
		// if only scaling up or down; but not out or in
		steps = append(steps, step.NewStepReplaceMaster(req.newNodeSize))
		// replace remaining replicas with resized nodes
		for i := 1; i < existingNodeCount; i++ {
			steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
		}
	} else {
		// changing both horizontal and vertical aspects of cluster
		steps = append(steps, step.NewStepReplaceMaster(req.newNodeSize))
		if req.isScalingOut() {
			for i := 1; i < existingNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := existingNodeCount; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.backends, req.logger))
			}
		} else if req.isScalingIn() {
			for i := 1; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := existingNodeCount; i > req.newNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.backends, req.logger))
			}
		}
	}
	return steps
}

// isInitialProvision is true if this Request is to create the initial cluster
func (req Request) isInitialProvision() bool {
	return req.cluster.MetaData().TargetNodeCount == 0
}

// isScalingUp is true if smaller nodes requested
func (req Request) isScalingUp() bool {
	return req.newNodeSize != 0 && defaultNodeSize < req.newNodeSize
}

// isScalingDown is true if bigger nodes requested
func (req Request) isScalingDown() bool {
	return req.newNodeSize != 0 && defaultNodeSize > req.newNodeSize
}

// isScalingOut is true if more nodes requested
func (req Request) isScalingOut() bool {
	return req.newNodeCount != 0 && req.cluster.MetaData().TargetNodeCount < req.newNodeCount
}

// isScalingIn is true if fewer nodes requested
func (req Request) isScalingIn() bool {
	return req.newNodeCount != 0 && req.cluster.MetaData().TargetNodeCount > req.newNodeCount
}

// logRequest send the requested change to Cluster to logs
func (req Request) logRequest() {
	req.logger.Info("request", lager.Data{
		"current-node-count": req.cluster.MetaData().TargetNodeCount,
		"new-node-count":     req.newNodeCount,
		"steps":              req.stepTypes(),
	})
}
