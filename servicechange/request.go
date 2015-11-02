package servicechange

// Request containers operations to perform a user-originating request to change a service instance (grow, scale, move)
type Request interface {
	Steps() []Step
}

// RealRequest represents a user-originating request to change a service instance (grow, scale, move)
type RealRequest struct {
}

// NewRequest creates a RealRequest to change a service instance
func NewRequest() Request {
	return RealRequest{}
}

// Steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (req RealRequest) Steps() []Step {
	return []Step{}
}
