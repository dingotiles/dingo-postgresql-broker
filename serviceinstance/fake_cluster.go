package serviceinstance

import "github.com/frodenas/brokerapi"

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

// ServiceDetails is the attributes of the Service Instance from Cloud Foundry
func (cluster FakeCluster) ServiceDetails() brokerapi.ProvisionDetails {
	return brokerapi.ProvisionDetails{}
}
