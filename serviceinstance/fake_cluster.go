package serviceinstance

// FakeCluster models a fake cluster of nodes
type FakeCluster struct {
	nodeCount int
	nodeSize  int
}

// NewFakeCluster creates a fake Cluster
func NewFakeCluster(nodeCount int, nodeSize int) FakeCluster {
	return FakeCluster{nodeCount: nodeCount, nodeSize: nodeSize}
}

// NodeCount is the total number of nodes in the cluster
func (cluster FakeCluster) NodeCount() int {
	return cluster.nodeCount
}

// NodeSize is the relative size of each node
func (cluster FakeCluster) NodeSize() int {
	return cluster.nodeSize
}
