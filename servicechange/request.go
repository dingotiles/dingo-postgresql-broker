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
	Cluster         serviceinstance.Cluster
	ChangeNodeSize  string
	ChangeNodeCount int
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
			for i := 0; i < req.ChangeNodeCount; i++ {
				step := NewStepAddNode()
				steps = append(steps, step)
			}
		}
		if req.IsScalingIn() {
			for i := 0; i < -req.ChangeNodeCount; i++ {
				step := NewStepRemoveNode()
				steps = append(steps, step)
			}
		}
	}
	return steps
}

// IsScalingUp is true if ChangeNodeSize is larger than current
// TODO: implement concept of node size
func (req RealRequest) IsScalingUp() bool {
	return false
}

// IsScalingDown is true if ChangeNodeSize is smaller than current
// TODO: implement concept of node size
func (req RealRequest) IsScalingDown() bool {
	return false
}

// IsScalingOut is true if ChangeNodeCount is positive
func (req RealRequest) IsScalingOut() bool {
	return req.ChangeNodeCount > 0
}

// IsScalingIn is true if ChangeNodeCount is negative
func (req RealRequest) IsScalingIn() bool {
	return req.ChangeNodeCount < 0
}
