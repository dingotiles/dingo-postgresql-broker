package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

// Request represents a user-originating request to change a service instance (grow, scale, move)
type Request struct {
	clusterState     structs.ClusterState
	features         structs.ClusterFeatures
	cluster          *state.Cluster
	backends         backend.Backends
	newNodeSize      int
	targetNodeCount  int
	currentNodeCount int
	logger           lager.Logger
}

// NewRequest creates a Request to change a service instance
func (s *Scheduler) NewRequest(cluster *state.Cluster) Request {
	return Request{
		currentNodeCount: len(cluster.AllNodes()),
		targetNodeCount:  cluster.MetaData().TargetNodeCount,
		cluster:          cluster,
		backends:         s.backends,
		newNodeSize:      defaultNodeSize,
		logger:           s.logger,
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
	existingNodeSize := defaultNodeSize
	targetNodeCount := req.targetNodeCount
	currentNodeCount := req.currentNodeCount

	steps := []step.Step{}
	if targetNodeCount == 0 {
		for i := currentNodeCount; i > targetNodeCount; i-- {
			steps = append(steps, step.NewStepRemoveNode(req.cluster, req.backends, req.logger))
		}
	} else if !req.isScalingUp() && !req.isScalingDown() &&
		!req.isScalingIn() && !req.isScalingOut() {
		return steps
	} else if !req.isScalingUp() && !req.isScalingDown() {
		// if only scaling out or in; but not up or down
		if req.isScalingOut() {
			for i := currentNodeCount; i < targetNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.backends, req.logger))
			}
		}
		if req.isScalingIn() {
			for i := currentNodeCount; i > targetNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.backends, req.logger))
			}
		}
	} else if !req.isScalingIn() && !req.isScalingOut() {
		// if only scaling up or down; but not out or in
		steps = append(steps, step.NewStepReplaceMaster(req.newNodeSize))
		// replace remaining replicas with resized nodes
		for i := 1; i < currentNodeCount; i++ {
			steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
		}
	} else {
		// changing both horizontal and vertical aspects of cluster
		steps = append(steps, step.NewStepReplaceMaster(req.newNodeSize))
		if req.isScalingOut() {
			for i := 1; i < currentNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := currentNodeCount; i < targetNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.backends, req.logger))
			}
		} else if req.isScalingIn() {
			for i := 1; i < targetNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := currentNodeCount; i > targetNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.backends, req.logger))
			}
		}
	}
	return steps
}

// isInitialProvision is true if this Request is to create the initial cluster
// func (req Request) isInitialProvision() bool {
// 	return req.cluster.MetaData().TargetNodeCount == 0
// }

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
	return req.targetNodeCount != 0 && req.currentNodeCount < req.targetNodeCount
}

// isScalingIn is true if fewer nodes requested
func (req Request) isScalingIn() bool {
	return req.targetNodeCount != 0 && req.currentNodeCount > req.targetNodeCount
}

// logRequest send the requested change to Cluster to logs
func (req Request) logRequest() {
	req.logger.Info("request", lager.Data{
		"current-node-count": req.currentNodeCount,
		"new-node-count":     req.targetNodeCount,
		"steps":              req.stepTypes(),
	})
}
