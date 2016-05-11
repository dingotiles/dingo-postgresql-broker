package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/pivotal-golang/lager"
)

const (
	defaultNodeSize     = 20
	debugBackendTraffic = false
)

// Request containers operations to perform a user-originating request to change a service instance (grow, scale, move)
type Request interface {
	// IsInitialProvision is true if this Request is to create the initial cluster
	IsInitialProvision() bool

	// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
	Steps() []step.Step

	// StepTypes is the ordered sequence of workflow step types to orchestrate a service instance change
	StepTypes() []string

	// IsScalingUp is true if nodes will grow in size
	IsScalingUp() bool

	// IsScalingUp is true if nodes will shrink in size
	IsScalingDown() bool

	// IsScalingOut is true if number of nodes will increase
	IsScalingOut() bool

	// IsScalingIn is true if number of nodes will decrease
	IsScalingIn() bool

	// Perform schedules the Request Steps() to be performed
	Perform() error
}

// RealRequest represents a user-originating request to change a service instance (grow, scale, move)
type RealRequest struct {
	cluster      *cluster.Cluster
	newNodeSize  int
	newNodeCount int
	logger       lager.Logger
}

// NewRequest creates a RealRequest to change a service instance
func NewRequest(cluster *cluster.Cluster, nodeCount int, logger lager.Logger) Request {
	return RealRequest{
		cluster:      cluster,
		newNodeCount: nodeCount,
		newNodeSize:  defaultNodeSize,
		logger:       logger,
	}
}

// StepTypes is the ordered sequence of workflow step types to orchestrate a service instance change
func (req RealRequest) StepTypes() []string {
	steps := req.Steps()
	stepTypes := make([]string, len(steps))
	for i, step := range steps {
		stepTypes[i] = step.StepType()
	}
	return stepTypes
}

// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (req RealRequest) Steps() []step.Step {
	existingNodeCount := req.cluster.Data.NodeCount
	existingNodeSize := defaultNodeSize
	steps := []step.Step{}
	if req.newNodeCount == 0 {
		for i := existingNodeCount; i > req.newNodeCount; i-- {
			steps = append(steps, step.NewStepRemoveNode(req.cluster, req.logger, debugBackendTraffic))
		}
	} else if !req.IsScalingUp() && !req.IsScalingDown() &&
		!req.IsScalingIn() && !req.IsScalingOut() {
		return steps
	} else if req.IsInitialProvision() {
		steps = append(steps, step.NewStepInitCluster(req.cluster, req.logger))
		for i := existingNodeCount; i < req.newNodeCount; i++ {
			steps = append(steps, step.NewStepAddNode(req.cluster, req.logger, debugBackendTraffic))
		}
	} else if !req.IsScalingUp() && !req.IsScalingDown() {
		// if only scaling out or in; but not up or down
		if req.IsScalingOut() {
			for i := existingNodeCount; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.cluster, req.logger, debugBackendTraffic))
			}
		}
		if req.IsScalingIn() {
			for i := existingNodeCount; i > req.newNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.logger, debugBackendTraffic))
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
				steps = append(steps, step.NewStepAddNode(req.cluster, req.logger, debugBackendTraffic))
			}
		} else if req.IsScalingIn() {
			for i := 1; i < req.newNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.newNodeSize))
			}
			for i := existingNodeCount; i > req.newNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.cluster, req.logger, debugBackendTraffic))
			}
		}
	}
	return steps
}

// IsInitialProvision is true if this Request is to create the initial cluster
func (req RealRequest) IsInitialProvision() bool {
	return req.cluster.Data.NodeCount == 0
}

// IsScalingUp is true if smaller nodes requested
func (req RealRequest) IsScalingUp() bool {
	return req.newNodeSize != 0 && defaultNodeSize < req.newNodeSize
}

// IsScalingDown is true if bigger nodes requested
func (req RealRequest) IsScalingDown() bool {
	return req.newNodeSize != 0 && defaultNodeSize > req.newNodeSize
}

// IsScalingOut is true if more nodes requested
func (req RealRequest) IsScalingOut() bool {
	return req.newNodeCount != 0 && req.cluster.Data.NodeCount < req.newNodeCount
}

// IsScalingIn is true if fewer nodes requested
func (req RealRequest) IsScalingIn() bool {
	return req.newNodeCount != 0 && req.cluster.Data.NodeCount > req.newNodeCount
}

// Perform schedules the Request Steps() to be performed
func (req RealRequest) Perform() (err error) {
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
func (req RealRequest) logRequest() {
	req.logger.Info("request", lager.Data{
		"current-node-count": req.cluster.Data.NodeCount,
		"new-node-count":     req.newNodeCount,
		"steps":              req.StepTypes(),
	})
}
