package serviceinstance

import (
	"fmt"
	"time"

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

// WaitForRoutingPortAllocation blocks until the routing tier has allocated a public port
func (cluster *Cluster) WaitForRoutingPortAllocation() (err error) {
	logger := cluster.Logger

	for index := 0; index < 10; index++ {
		key := fmt.Sprintf("/routing/allocation/%s", cluster.InstanceID)
		resp, err := cluster.EtcdClient.Get(key, false, false)
		if err != nil {
			logger.Debug("cluster.provision.routing", lager.Data{"polling": "allocated-port"})
		} else {
			logger.Info("cluster.provision..routing", lager.Data{"allocated-port": resp.Node.Value})
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	logger.Error("cluster.provision.routing", err)
	return err
}
