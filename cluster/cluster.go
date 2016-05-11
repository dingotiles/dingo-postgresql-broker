package cluster

import (
	"fmt"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/bkrconfig"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Cluster describes a real/proposed cluster of nodes
type Cluster struct {
	config     *bkrconfig.Config
	etcdClient backend.EtcdClient
	logger     lager.Logger
	Data       ClusterData
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromProvisionDetails(instanceID string, details brokerapi.ProvisionDetails, etcdClient backend.EtcdClient, config *bkrconfig.Config, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		etcdClient: etcdClient,
		config:     config,
		Data: ClusterData{
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
			"instance-id": cluster.Data.InstanceID,
			"service-id":  cluster.Data.ServiceID,
			"plan-id":     cluster.Data.PlanID,
		})
	}
	return
}

// NewCluster creates a RealCluster from ProvisionDetails
func NewClusterFromRestoredData(instanceID string, clusterdata *ClusterData, etcdClient backend.EtcdClient, config *bkrconfig.Config, logger lager.Logger) (cluster *Cluster) {
	cluster = &Cluster{
		etcdClient: etcdClient,
		config:     config,
		Data:       *clusterdata,
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
func (cluster *Cluster) Exists() bool {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID)
	_, err := cluster.etcdClient.Get(key, false, true)
	return err == nil
}

// Load the cluster information from KV store
func (cluster *Cluster) Load() error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID)
	resp, err := cluster.etcdClient.Get(key, false, true)
	if err != nil {
		cluster.logger.Error("load.etcd-get", err)
		return err
	}
	cluster.Data.NodeCount = len(resp.Node.Nodes)
	cluster.logger.Info("load.state", lager.Data{
		"node-count": cluster.Data.NodeCount,
	})
	return nil
}

func (cluster *Cluster) Init() error {
	key := fmt.Sprintf("/serviceinstances/%s/plan_id", cluster.Data.InstanceID)
	_, err := cluster.etcdClient.Set(key, cluster.Data.PlanID, 0)
	return err
}

// WaitForRoutingPortAllocation blocks until the routing tier has allocated a public port
func (cluster *Cluster) WaitForRoutingPortAllocation() (err error) {
	for index := 0; index < 10; index++ {
		key := fmt.Sprintf("/routing/allocation/%s", cluster.Data.InstanceID)
		resp, err := cluster.etcdClient.Get(key, false, false)
		if err != nil {
			cluster.logger.Debug("provision.routing.polling", lager.Data{})
		} else {
			cluster.Data.AllocatedPort = resp.Node.Value
			cluster.logger.Info("provision.routing.done", lager.Data{"allocated_port": cluster.Data.AllocatedPort})
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	cluster.logger.Error("provision.routing.timed-out", err, lager.Data{"err": err})
	return err
}
