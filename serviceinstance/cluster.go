package serviceinstance

import (
	"fmt"
	"math/rand"
	"regexp"
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

// Load the cluster information from KV store
func (cluster *Cluster) Load() error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, true)
	if err != nil {
		cluster.Logger.Error("cluster.load", err)
		return err
	}
	cluster.NodeCount = len(resp.Node.Nodes)
	cluster.NodeSize = 20
	cluster.Logger.Info("cluster.load", lager.Data{"node-count": cluster.NodeCount, "node-size": 20})
	return nil
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
			logger.Info("cluster.provision.routing", lager.Data{"allocated-port": resp.Node.Value})
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	logger.Error("cluster.provision.routing", err)
	return err
}

// RandomReplicaNode should discover which nodes are replicas and return a random one
// FIXME - currently just picking a random node - which might be the master
func (cluster *Cluster) RandomReplicaNode() (nodeUUID string, backend string, err error) {
	logger := cluster.Logger
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, true)
	if err != nil {
		logger.Error("cluster.random-replica-node.nodes", err)
		return
	}
	item := rand.Intn(len(resp.Node.Nodes))
	nodeKey := resp.Node.Nodes[item].Key
	r, _ := regexp.Compile("/nodes/(.*)$")
	matches := r.FindStringSubmatch(nodeKey)
	nodeUUID = matches[1]

	key = fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", cluster.InstanceID, nodeUUID)
	resp, err = cluster.EtcdClient.Get(key, false, false)
	if err != nil {
		logger.Error("cluster.random-replica-node.backend", err)
		return
	}
	backend = resp.Node.Value

	return
}
