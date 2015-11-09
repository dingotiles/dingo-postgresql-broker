package serviceinstance

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/utils"
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

// Currently hardcoded list of backends in demo cluster, with example AZ split
func (cluster *Cluster) allBackends() (backends []*backend.Backend) {
	return []*backend.Backend{
		&backend.Backend{AvailabilityZone: "z1", GUID: "10.244.21.6", URI: "http://54.145.50.109:10006", Username: "containers", Password: "containers"},
		&backend.Backend{AvailabilityZone: "z1", GUID: "10.244.21.7", URI: "http://54.145.50.109:10007", Username: "containers", Password: "containers"},
		&backend.Backend{AvailabilityZone: "z2", GUID: "10.244.21.8", URI: "http://54.145.50.109:10008", Username: "containers", Password: "containers"},
	}
	// list := rand.Perm(len(backends))
}

// List of AZs offered by allBackends()
// FIXME - currently hardcoded; should be built/cached from allBackends
func (cluster *Cluster) allAZs() []string {
	return []string{"z1", "z2"}
}

// if any errors, assume that cluster has no running nodes yet
func (cluster *Cluster) usedBackendGUIDs() (backendGUIDs []string) {
	logger := cluster.Logger
	resp, err := cluster.EtcdClient.Get(fmt.Sprintf("/serviceinstances/%s/nodes", cluster.InstanceID), false, false)
	if err != nil {
		return
	}
	for _, clusterNode := range resp.Node.Nodes {
		nodeKey := clusterNode.Key
		resp, err = cluster.EtcdClient.Get(fmt.Sprintf("%s/backend", nodeKey), false, false)
		if err != nil {
			logger.Error("cluster.az-used.backend", err)
			return
		}
		backendGUIDs = append(backendGUIDs, resp.Node.Value)
	}
	return
}

// backendAZsByUnusedness sorts the availability zones in order of whether this cluster is using them or not
// An AZ that is not being used at all will be early in the result.
// All known AZs are included in the result
func (cluster *Cluster) sortBackendAZsByUnusedness() (vs *utils.ValSorter) {
	backends := cluster.allBackends()
	azUsageData := map[string]int{}
	for _, az := range cluster.allAZs() {
		azUsageData[az] = 0
	}
	for _, backendGUID := range cluster.usedBackendGUIDs() {
		for _, backend := range backends {
			if backend.GUID == backendGUID {
				azUsageData[backend.AvailabilityZone]++
			}
		}
	}
	vs = utils.NewValSorter(azUsageData)
	vs.Sort()
	return
}

// SortedBackendsByUnusedAZs is sequence of backends to try to request new nodes for this cluster
// It prioritizes backends in availability zones that are not currently used
func (cluster *Cluster) SortedBackendsByUnusedAZs() (backends []*backend.Backend) {
	for _, az := range cluster.sortBackendAZsByUnusedness().Keys {
		for _, backend := range cluster.allBackends() {
			if backend.AvailabilityZone == az {
				backends = append(backends, backend)
			}
		}
	}
	return
}
