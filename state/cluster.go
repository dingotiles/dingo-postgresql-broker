package state

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Cluster describes a real/proposed cluster of nodes
type Cluster struct {
	etcdClient backend.EtcdClient
	logger     lager.Logger
	meta       structs.ClusterData
}

func (c *Cluster) MetaData() structs.ClusterData {
	return c.meta
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromProvisionDetails(instanceID string, details brokerapi.ProvisionDetails, etcdClient backend.EtcdClient, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		etcdClient: etcdClient,
		meta: structs.ClusterData{
			InstanceID:       instanceID,
			OrganizationGUID: details.OrganizationGUID,
			PlanID:           details.PlanID,
			ServiceID:        details.ServiceID,
			SpaceGUID:        details.SpaceGUID,
			AdminCredentials: structs.AdminCredentials{},
		},
	}
	if logger != nil {
		cluster.logger = logger.Session("cluster", lager.Data{
			"instance-id": cluster.MetaData().InstanceID,
			"service-id":  cluster.MetaData().ServiceID,
			"plan-id":     cluster.MetaData().PlanID,
		})
	}
	return
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromRestoredData(instanceID string, clusterdata *structs.ClusterData, etcdClient backend.EtcdClient, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		etcdClient: etcdClient,
		meta:       *clusterdata,
	}
	if logger != nil {
		cluster.logger = logger.Session("cluster", lager.Data{
			"instance-id": clusterdata.InstanceID,
			"service-id":  clusterdata.ServiceID,
			"plan-id":     clusterdata.PlanID,
		})
	}
	return
}

// Load the cluster information from KV store
func (cluster *Cluster) Load() error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.MetaData().InstanceID)
	resp, err := cluster.etcdClient.Get(key, false, true)
	if err != nil {
		cluster.logger.Error("load.etcd-get", err)
		return err
	}
	cluster.meta.TargetNodeCount = len(resp.Node.Nodes)
	cluster.logger.Info("load.state", lager.Data{
		"node-count": cluster.MetaData().TargetNodeCount,
	})
	return nil
}

// TODO write ClusterData to etcd
func (c *Cluster) writeState() error {
	c.logger.Info("write-state")
	key := fmt.Sprintf("/serviceinstances/%s/plan_id", c.meta.InstanceID)
	_, err := c.etcdClient.Set(key, c.meta.PlanID, 0)
	if err != nil {
		c.logger.Error("write-state.error", err)
		return err
	}
	return err
}

func (c *Cluster) PortAllocation() (int64, error) {
	key := fmt.Sprintf("/routing/allocation/%s", c.meta.InstanceID)
	resp, err := c.etcdClient.Get(key, false, false)
	if err != nil {
		c.logger.Error("routing-allocation.get", err)
		return 0, err
	}
	publicPort, err := strconv.ParseInt(resp.Node.Value, 10, 64)
	if err != nil {
		c.logger.Error("bind.routing-allocation.parse-int", err)
		return 0, err
	}
	return publicPort, nil
}

// WaitForRoutingPortAllocation blocks until the routing tier has allocated a public port
func (cluster *Cluster) WaitForRoutingPortAllocation() (err error) {
	for index := 0; index < 10; index++ {
		key := fmt.Sprintf("/routing/allocation/%s", cluster.MetaData().InstanceID)
		resp, err := cluster.etcdClient.Get(key, false, false)
		if err != nil {
			cluster.logger.Debug("provision.routing.polling", lager.Data{})
		} else {
			cluster.meta.AllocatedPort = resp.Node.Value
			cluster.logger.Info("provision.routing.done", lager.Data{"allocated_port": cluster.MetaData().AllocatedPort})
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	cluster.logger.Error("provision.routing.timed-out", err, lager.Data{"err": err})
	return err
}
