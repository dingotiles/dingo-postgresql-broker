package serviceinstance

import "github.com/frodenas/brokerapi"

// Cluster describes the operations performed upon a Cluster
type Cluster interface {
	// NodeCount is the total number of nodes in the cluster
	NodeCount() int
	// NodeSize is the relative size of each node
	NodeSize() int
	// ServiceDetails is the attributes of the Service Instance from Cloud Foundry
	ServiceDetails() brokerapi.ProvisionDetails
}

// RealCluster describes a real/proposed cluster of nodes
type RealCluster struct {
	instanceID     string
	nodeCount      int
	nodeSize       int
	serviceDetails brokerapi.ProvisionDetails
}

// NewCluster creates a RealCluster
func NewCluster(instanceID string, details brokerapi.ProvisionDetails) RealCluster {
	return RealCluster{instanceID: instanceID, serviceDetails: details}
}

// NodeCount is the total number of nodes in the cluster
func (cluster RealCluster) NodeCount() int {
	return cluster.nodeCount
}

// NodeSize is the relative size of each node
func (cluster RealCluster) NodeSize() int {
	return cluster.nodeSize
}

// ServiceDetails is the attributes of the Service Instance from Cloud Foundry
func (cluster RealCluster) ServiceDetails() brokerapi.ProvisionDetails {
	return cluster.serviceDetails
}
