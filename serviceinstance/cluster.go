package serviceinstance

// Cluster describes the operations performed upon a Cluster
type Cluster interface {
	// NodeCount is the total number of nodes in the cluster
	NodeCount() uint
	// NodeSize is the relative size of each node
	NodeSize() uint
}

// RealCluster describes a real/proposed cluster of nodes
type RealCluster struct {
	nodeCount uint
	nodeSize  uint
}

// NewCluster creates a RealCluster
func NewCluster() RealCluster {
	return RealCluster{}
}

// NodeCount is the total number of nodes in the cluster
func (cluster RealCluster) NodeCount() uint {
	return cluster.nodeCount
}

// NodeSize is the relative size of each node
func (cluster RealCluster) NodeSize() uint {
	return cluster.nodeSize
}
