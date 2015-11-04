package serviceinstance

// Cluster describes the operations performed upon a Cluster
type Cluster interface {
	// NodeCount is the total number of nodes in the cluster
	NodeCount() int
	// NodeSize is the relative size of each node
	NodeSize() int
}

// RealCluster describes a real/proposed cluster of nodes
type RealCluster struct {
	instanceID string
	nodeCount  int
	nodeSize   int
}

// NewCluster creates a RealCluster
func NewCluster(instanceID string) RealCluster {
	return RealCluster{instanceID: instanceID}
}

// NodeCount is the total number of nodes in the cluster
func (cluster RealCluster) NodeCount() int {
	return cluster.nodeCount
}

// NodeSize is the relative size of each node
func (cluster RealCluster) NodeSize() int {
	return cluster.nodeSize
}
