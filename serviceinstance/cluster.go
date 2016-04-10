package serviceinstance

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Cluster describes a real/proposed cluster of nodes
type Cluster struct {
	Config     *config.Config
	EtcdClient backend.EtcdClient
	Logger     lager.Logger
	Data       ClusterData
}

// ClusterData describes the current request for the state of the cluster
type ClusterData struct {
	InstanceID       string                 `json:"instance_id"`
	ServiceID        string                 `json:"service_id"`
	PlanID           string                 `json:"plan_id"`
	OrganizationGUID string                 `json:"organization_guid"`
	SpaceGUID        string                 `json:"space_guid"`
	Parameters       map[string]interface{} `json:"parameters"`
	NodeCount        int                    `json:"node_count"`
	NodeSize         int                    `json:"node_size"`
	AllocatedPort    string                 `json:"allocated_port"`
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromProvisionDetails(instanceID string, details brokerapi.ProvisionDetails, etcdClient backend.EtcdClient, config *config.Config, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		EtcdClient: etcdClient,
		Config:     config,
		Data: ClusterData{
			InstanceID:       instanceID,
			OrganizationGUID: details.OrganizationGUID,
			PlanID:           details.PlanID,
			ServiceID:        details.ServiceID,
			SpaceGUID:        details.SpaceGUID,
			Parameters:       details.Parameters,
		},
	}
	if logger != nil {
		cluster.Logger = logger.Session("cluster", lager.Data{
			"instance-id": cluster.Data.InstanceID,
			"service-id":  cluster.Data.ServiceID,
			"plan-id":     cluster.Data.PlanID,
		})
	}
	return
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromRestoredData(instanceID string, clusterdata *ClusterData, etcdClient backend.EtcdClient, config *config.Config, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		EtcdClient: etcdClient,
		Config:     config,
		Data:       *clusterdata,
	}
	if logger != nil {
		cluster.Logger = logger.Session("cluster", lager.Data{
			"instance-id": clusterdata.InstanceID,
			"service-id":  clusterdata.ServiceID,
			"plan-id":     clusterdata.PlanID,
		})
	}
	return
}

// ClusterData describes the current request for the state of the cluster
func (cluster *Cluster) ClusterData() *ClusterData {
	return &cluster.Data
}

// Exists returns true if cluster already exists
func (cluster *Cluster) Exists() bool {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID)
	_, err := cluster.EtcdClient.Get(key, false, true)
	return err == nil
}

// Load the cluster information from KV store
func (cluster *Cluster) Load() error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, true)
	if err != nil {
		cluster.Logger.Error("load.etcd-get", err)
		return err
	}
	cluster.Data.NodeCount = len(resp.Node.Nodes)
	// TODO load current node size
	cluster.Data.NodeSize = 20
	cluster.Logger.Info("load.state", lager.Data{
		"node-count": cluster.Data.NodeCount,
		"node-size":  cluster.Data.NodeSize,
	})
	return nil
}

// WaitForRoutingPortAllocation blocks until the routing tier has allocated a public port
func (cluster *Cluster) WaitForRoutingPortAllocation() (err error) {
	for index := 0; index < 10; index++ {
		key := fmt.Sprintf("/routing/allocation/%s", cluster.Data.InstanceID)
		resp, err := cluster.EtcdClient.Get(key, false, false)
		if err != nil {
			cluster.Logger.Debug("provision.routing.polling", lager.Data{})
		} else {
			cluster.Data.AllocatedPort = resp.Node.Value
			cluster.Logger.Info("provision.routing.done", lager.Data{"allocated_port": cluster.Data.AllocatedPort})
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	cluster.Logger.Error("provision.routing.timed-out", err, lager.Data{"err": err})
	return err
}

// RandomReplicaNode should discover which nodes are replicas and return a random one
// FIXME - currently just picking a random node - which might be the master
func (cluster *Cluster) RandomReplicaNode() (nodeUUID string, backend string, err error) {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, true)
	if err != nil {
		cluster.Logger.Error("random-replica-node.nodes", err)
		return
	}
	item := rand.Intn(len(resp.Node.Nodes))
	nodeKey := resp.Node.Nodes[item].Key
	r, _ := regexp.Compile("/nodes/(.*)$")
	matches := r.FindStringSubmatch(nodeKey)
	nodeUUID = matches[1]

	key = fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", cluster.Data.InstanceID, nodeUUID)
	resp, err = cluster.EtcdClient.Get(key, false, false)
	if err != nil {
		cluster.Logger.Error("random-replica-node.backend", err)
		return
	}
	backend = resp.Node.Value

	return
}

// AllBackends is a flat list of all Backend APIs
func (cluster *Cluster) AllBackends() (backends []*config.Backend) {
	return cluster.Config.Backends
}

// AllAZs lists of AZs offered by AllBackends()
func (cluster *Cluster) AllAZs() (list []string) {
	azUsage := map[string]int{}
	for _, backend := range cluster.AllBackends() {
		azUsage[backend.AvailabilityZone]++
	}
	for az := range azUsage {
		list = append(list, az)
	}
	// TEST sorting AZs for benefit of tests
	sort.Strings(list)
	return
}

// if any errors, assume that cluster has no running nodes yet
func (cluster *Cluster) usedBackendGUIDs() (backendGUIDs []string) {
	resp, err := cluster.EtcdClient.Get(fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID), false, false)
	if err != nil {
		return
	}
	for _, clusterNode := range resp.Node.Nodes {
		nodeKey := clusterNode.Key
		resp, err = cluster.EtcdClient.Get(fmt.Sprintf("%s/backend", nodeKey), false, false)
		if err != nil {
			cluster.Logger.Error("az-used.backend", err)
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
	backends := cluster.AllBackends()
	azUsageData := map[string]int{}
	for _, az := range cluster.AllAZs() {
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
	fmt.Printf("usage %#v\n", azUsageData)
	vs.Sort()
	fmt.Printf("sorted %#v\n", vs)
	return
}
