package servicechange

import "github.com/cloudfoundry-community/patroni-broker/serviceinstance"

// Request containers operations to perform a user-originating request to change a service instance (grow, scale, move)
type Request interface {
	Steps() []Step
}

// RealRequest represents a user-originating request to change a service instance (grow, scale, move)
type RealRequest struct {
	Cluster         serviceinstance.Cluster
	ChangeNodeCount int
	ChangeNodeSize  string
}

// NewRequest creates a RealRequest to change a service instance
func NewRequest(cluster serviceinstance.Cluster) RealRequest {
	return RealRequest{Cluster: cluster}
}

// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (req RealRequest) Steps() []Step {
	return []Step{}
}
