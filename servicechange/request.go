package servicechange

import (
	"github.com/dingotiles/dingo-postgresql-broker/servicechange/step"
	"github.com/dingotiles/dingo-postgresql-broker/serviceinstance"
	"github.com/pivotal-golang/lager"
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
	Cluster      *serviceinstance.Cluster
	NewNodeSize  int
	NewNodeCount int
}

// NewRequest creates a RealRequest to change a service instance
func NewRequest(cluster *serviceinstance.Cluster, nodeCount, nodeSize int) Request {
	return RealRequest{
		Cluster:      cluster,
		NewNodeCount: nodeCount,
		NewNodeSize:  nodeSize,
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
	existingNodeCount := req.Cluster.Data.NodeCount
	existingNodeSize := req.Cluster.Data.NodeSize
	steps := []step.Step{}
	if req.NewNodeCount == 0 {
		for i := existingNodeCount; i > req.NewNodeCount; i-- {
			steps = append(steps, step.NewStepRemoveNode(req.Cluster))
		}
	} else if !req.IsScalingUp() && !req.IsScalingDown() &&
		!req.IsScalingIn() && !req.IsScalingOut() {
		return steps
	} else if req.IsInitialProvision() {
		steps = append(steps, step.NewStepInitCluster(req.Cluster))
		for i := existingNodeCount; i < req.NewNodeCount; i++ {
			steps = append(steps, step.NewStepAddNode(req.Cluster))
		}
	} else if !req.IsScalingUp() && !req.IsScalingDown() {
		// if only scaling out or in; but not up or down
		if req.IsScalingOut() {
			for i := existingNodeCount; i < req.NewNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.Cluster))
			}
		}
		if req.IsScalingIn() {
			for i := existingNodeCount; i > req.NewNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.Cluster))
			}
		}
	} else if !req.IsScalingIn() && !req.IsScalingOut() {
		// if only scaling up or down; but not out or in
		steps = append(steps, step.NewStepReplaceMaster(req.NewNodeSize))
		// replace remaining replicas with resized nodes
		for i := 1; i < existingNodeCount; i++ {
			steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.NewNodeSize))
		}
	} else {
		// changing both horizontal and vertical aspects of cluster
		steps = append(steps, step.NewStepReplaceMaster(req.NewNodeSize))
		if req.IsScalingOut() {
			for i := 1; i < existingNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.NewNodeSize))
			}
			for i := existingNodeCount; i < req.NewNodeCount; i++ {
				steps = append(steps, step.NewStepAddNode(req.Cluster))
			}
		} else if req.IsScalingIn() {
			for i := 1; i < req.NewNodeCount; i++ {
				steps = append(steps, step.NewStepReplaceReplica(existingNodeSize, req.NewNodeSize))
			}
			for i := existingNodeCount; i > req.NewNodeCount; i-- {
				steps = append(steps, step.NewStepRemoveNode(req.Cluster))
			}
		}
	}
	return steps
}

// IsInitialProvision is true if this Request is to create the initial cluster
func (req RealRequest) IsInitialProvision() bool {
	return req.Cluster.Data.NodeCount == 0
}

// IsScalingUp is true if smaller nodes requested
func (req RealRequest) IsScalingUp() bool {
	return req.NewNodeSize != 0 && req.Cluster.Data.NodeSize < req.NewNodeSize
}

// IsScalingDown is true if bigger nodes requested
func (req RealRequest) IsScalingDown() bool {
	return req.NewNodeSize != 0 && req.Cluster.Data.NodeSize > req.NewNodeSize
}

// IsScalingOut is true if more nodes requested
func (req RealRequest) IsScalingOut() bool {
	return req.NewNodeCount != 0 && req.Cluster.Data.NodeCount < req.NewNodeCount
}

// IsScalingIn is true if fewer nodes requested
func (req RealRequest) IsScalingIn() bool {
	return req.NewNodeCount != 0 && req.Cluster.Data.NodeCount > req.NewNodeCount
}

// Perform schedules the Request Steps() to be performed
func (req RealRequest) Perform() (err error) {
	req.logRequest()
	if len(req.Steps()) == 0 {
		req.logger().Info("request.no-steps")
		return
	}
	req.logger().Info("request.perform", lager.Data{"steps-count": len(req.Steps())})
	for _, step := range req.Steps() {
		err = step.Perform()
		if err != nil {
			return
		}
	}
	return
}

func (req RealRequest) logger() lager.Logger {
	return req.Cluster.Logger
}

// logRequest send the requested change to Cluster to logs
func (req RealRequest) logRequest() {
	req.logger().Info("request", lager.Data{
		"current-node-count": req.Cluster.Data.NodeCount,
		"current-node-size":  req.Cluster.Data.NodeSize,
		"new-node-count":     req.NewNodeCount,
		"new-node-size":      req.NewNodeSize,
		"steps":              req.StepTypes(),
	})
}
