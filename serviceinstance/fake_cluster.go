package serviceinstance

// FakeCluster models a fake cluster of nodes
type FakeCluster struct {
	nodeCount uint
	nodeSize  uint
}

// NewFakeCluster creates a fake Cluster
func NewFakeCluster(nodeCount uint, nodeSize uint) FakeCluster {
	return FakeCluster{nodeCount: nodeCount, nodeSize: nodeSize}
}

// NodeCount is the total number of nodes in the cluster
func (cluster FakeCluster) NodeCount() uint {
	return cluster.nodeCount
}

// NodeSize is the relative size of each node
func (cluster FakeCluster) NodeSize() uint {
	return cluster.nodeSize
}
