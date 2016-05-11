package state

import (
	"fmt"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Cluster describes a real/proposed cluster of nodes
type Cluster struct {
	config     *config.Config
	etcdClient backend.EtcdClient
	logger     lager.Logger
	meta       MetaData
}

func (c *Cluster) MetaData() MetaData {
	return c.meta
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromProvisionDetails(instanceID string, details brokerapi.ProvisionDetails, etcdClient backend.EtcdClient, config *config.Config, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		etcdClient: etcdClient,
		config:     config,
		meta: MetaData{
			InstanceID:       instanceID,
			OrganizationGUID: details.OrganizationGUID,
			PlanID:           details.PlanID,
			ServiceID:        details.ServiceID,
			SpaceGUID:        details.SpaceGUID,
			AdminCredentials: AdminCredentials{
				Username: "pgadmin",
				Password: NewPassword(16),
			},
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
func NewClusterFromRestoredData(instanceID string, clusterdata *MetaData, etcdClient backend.EtcdClient, config *config.Config, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		etcdClient: etcdClient,
		config:     config,
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

// Exists returns true if cluster already exists
func Exists(etcdClient backend.EtcdClient, instanceId string) bool {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", instanceId)
	_, err := etcdClient.Get(key, false, true)
	return err == nil
}

// Load the cluster information from KV store
func (cluster *Cluster) Load() error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.MetaData().InstanceID)
	resp, err := cluster.etcdClient.Get(key, false, true)
	if err != nil {
		cluster.logger.Error("load.etcd-get", err)
		return err
	}
	cluster.meta.NodeCount = len(resp.Node.Nodes)
	cluster.logger.Info("load.state", lager.Data{
		"node-count": cluster.MetaData().NodeCount,
	})
	return nil
}

func (cluster *Cluster) Init() error {
	key := fmt.Sprintf("/serviceinstances/%s/plan_id", cluster.MetaData().InstanceID)
	_, err := cluster.etcdClient.Set(key, cluster.MetaData().PlanID, 0)
	return err
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
