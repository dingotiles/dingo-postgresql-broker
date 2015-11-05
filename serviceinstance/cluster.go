package serviceinstance

import (
	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Cluster describes a real/proposed cluster of nodes
type Cluster struct {
	EtcdClient     *backend.EtcdClient
	Logger         lager.Logger
	InstanceID     string
	NodeCount      int
	NodeSize       int
	ServiceDetails brokerapi.ProvisionDetails
}

// NewCluster creates a RealCluster
func NewCluster(instanceID string, details brokerapi.ProvisionDetails, etcdClient *backend.EtcdClient, logger lager.Logger) *Cluster {
	return &Cluster{
		InstanceID:     instanceID,
		ServiceDetails: details,
		EtcdClient:     etcdClient,
		Logger:         logger,
	}
}
