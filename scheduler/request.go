package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/pivotal-golang/lager"
)

const (
	defaultNodeSize = 20
)

// Request represents a user-originating request to change a service instance (grow, scale, move)
type Request struct {
	cluster      *cluster.Cluster
	newNodeSize  int
	newNodeCount int
	logger       lager.Logger
}

// NewRequest creates a Request to change a service instance
func NewRequest(cluster *cluster.Cluster, nodeCount int, logger lager.Logger) Request {
	return Request{
		cluster:      cluster,
		newNodeCount: nodeCount,
		newNodeSize:  defaultNodeSize,
		logger:       logger,
	}
}

// StepTypes is the ordered sequence of workflow step types to orchestrate a service instance change
func (req Request) StepTypes() []string {
	steps := req.Steps()
	stepTypes := make([]string, len(steps))
	for i, step := range steps {
		stepTypes[i] = step.StepType()
	}
	return stepTypes
}

// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (req Request) Steps() []step.Step {
	existingNodeCount := req.cluster.Data.NodeCount
	existingNodeSize := defaultNodeSize
	steps := []step.Step{}
	if req.newNodeCount == 0 {
		for i := existingNodeCount; i > req.newNodeCount; i-- {
			steps = append(steps, step.NewStepRemoveNode(req.cluster, req.logger))
		}
	} else if !req.IsScalingUp() && !req.IsScalingDown() &&
		!req.IsScalingIn() && !req.IsScalingOut() {
		return steps
	} else if req.IsInitialProvision() {
		steps = append(steps, step.NewStepInitCluster(req.cluster, req.logger))
		for i := existingNodeCount; i < req.newNodeCount; i++ {
			steps = append(steps, step.NewStepAddNode(req.cluster, req.logger))
		}
	} else if !req.IsScalingUp() && !req.IsScalingDown() {
		// if only scaling out or in; but not up or down
		if req.IsScalingOut() {
			for i := existingNodeCount; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.logger))
			}
		}
		if req.IsScalingIn() {
			for i := existingNodeCount; i > req.newNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.logger))
			}
		}
	} else if !req.IsScalingIn() && !req.IsScalingOut() {
		// if only scaling up or down; but not out or in
		steps = append(steps, step.NewStepReplaceMaster(req.newNodeSize))
		// replace remaining replicas with resized nodes
		for i := 1; i < existingNodeCount; i++ {
			steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
		}
	} else {
		// changing both horizontal and vertical aspects of cluster
		steps = append(steps, step.NewStepReplaceMaster(req.newNodeSize))
		if req.IsScalingOut() {
			for i := 1; i < existingNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := existingNodeCount; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.logger))
			}
		} else if req.IsScalingIn() {
			for i := 1; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := existingNodeCount; i > req.newNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.logger))
			}
		}
	}
	return steps
}

// IsInitialProvision is true if this Request is to create the initial cluster
func (req Request) IsInitialProvision() bool {
	return req.cluster.Data.NodeCount == 0
}

// IsScalingUp is true if smaller nodes requested
func (req Request) IsScalingUp() bool {
	return req.newNodeSize != 0 && defaultNodeSize < req.newNodeSize
}

// IsScalingDown is true if bigger nodes requested
func (req Request) IsScalingDown() bool {
	return req.newNodeSize != 0 && defaultNodeSize > req.newNodeSize
}

// IsScalingOut is true if more nodes requested
func (req Request) IsScalingOut() bool {
	return req.newNodeCount != 0 && req.cluster.Data.NodeCount < req.newNodeCount
}

// IsScalingIn is true if fewer nodes requested
func (req Request) IsScalingIn() bool {
	return req.newNodeCount != 0 && req.cluster.Data.NodeCount > req.newNodeCount
}

// Perform schedules the Request Steps() to be performed
func (req Request) Perform() (err error) {
	req.logRequest()
	if len(req.Steps()) == 0 {
		req.logger.Info("request.no-steps")
		return
	}
	req.logger.Info("request.perform", lager.Data{"steps-count": len(req.Steps())})
	for _, step := range req.Steps() {
		err = step.Perform()
		if err != nil {
			return
		}
	}
	return
}

// logRequest send the requested change to Cluster to logs
func (req Request) logRequest() {
	req.logger.Info("request", lager.Data{
		"current-node-count": req.cluster.Data.NodeCount,
		"new-node-count":     req.newNodeCount,
		"steps":              req.StepTypes(),
	})
}
