package servicechange

import "github.com/cloudfoundry-community/patroni-broker/serviceinstance"

// Request containers operations to perform a user-originating request to change a service instance (grow, scale, move)
type Request interface {
	// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
	Steps() []Step

	// IsScalingUp is true if nodes will grow in size
	IsScalingUp() bool

	// IsScalingUp is true if nodes will shrink in size
	IsScalingDown() bool

	// IsScalingOut is true if number of nodes will increase
	IsScalingOut() bool

	// IsScalingIn is true if number of nodes will decrease
	IsScalingIn() bool
}

// RealRequest represents a user-originating request to change a service instance (grow, scale, move)
type RealRequest struct {
	Cluster      serviceinstance.Cluster
	NewNodeSize  uint
	NewNodeCount uint
}

// NewRequest creates a RealRequest to change a service instance
func NewRequest(cluster serviceinstance.Cluster) RealRequest {
	return RealRequest{Cluster: cluster}
}

// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (req RealRequest) Steps() []Step {
	steps := []Step{}
	if !req.IsScalingUp() && !req.IsScalingDown() {
		if req.IsScalingOut() {
			for i := req.Cluster.NodeCount(); i < req.NewNodeCount; i++ {
				step := NewStepAddNode()
				steps = append(steps, step)
			}
		}
		if req.IsScalingIn() {
			for i := req.Cluster.NodeCount(); i > req.NewNodeCount; i-- {
				step := NewStepRemoveNode()
				steps = append(steps, step)
			}
		}
	}
	return steps
}

// IsScalingUp is true if smaller nodes requested
func (req RealRequest) IsScalingUp() bool {
	return req.NewNodeSize != 0 && req.Cluster.NodeSize() < req.NewNodeSize
}

// IsScalingDown is true if bigger nodes requested
func (req RealRequest) IsScalingDown() bool {
	return req.NewNodeSize != 0 && req.Cluster.NodeSize() > req.NewNodeSize
}

// IsScalingOut is true if more nodes requested
func (req RealRequest) IsScalingOut() bool {
	return req.NewNodeCount != 0 && req.Cluster.NodeCount() < req.NewNodeCount
}

// IsScalingIn is true if fewer nodes requested
func (req RealRequest) IsScalingIn() bool {
	return req.NewNodeCount != 0 && req.Cluster.NodeCount() > req.NewNodeCount
}
