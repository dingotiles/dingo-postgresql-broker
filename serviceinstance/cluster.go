package serviceinstance

// Cluster describes the operations performed upon a Cluster
type Cluster interface {
}

// RealCluster describes a real/proposed cluster of nodes
type RealCluster struct {
}

// NewCluster creates a RealCluster
func NewCluster() RealCluster {
	return RealCluster{}
}
