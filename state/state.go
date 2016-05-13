package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/pivotal-golang/lager"
)

type State interface {

	// ClusterExists returns true if cluster already exists
	ClusterExists(clusterID string) bool
	InitializeCluster(clusterData *structs.ClusterData) (*Cluster, error)
}

type etcdState struct {
	etcd   backend.EtcdClient
	logger lager.Logger
}

func NewState(etcdClient backend.EtcdClient, logger lager.Logger) State {
	return &etcdState{
		etcd:   etcdClient,
		logger: logger,
	}
}

func (s *etcdState) ClusterExists(instanceID string) bool {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", instanceID)
	_, err := s.etcd.Get(key, false, true)
	return err == nil
}

func (s *etcdState) InitializeCluster(clusterData *structs.ClusterData) (*Cluster, error) {
	cluster := &Cluster{
		etcdClient: s.etcd,
		logger: s.logger.Session("cluster", lager.Data{
			"instance-id": clusterData.InstanceID,
			"service-id":  clusterData.ServiceID,
			"plan-id":     clusterData.PlanID,
		}),
		meta: *clusterData,
	}
	err := cluster.writeState()
	if err != nil {
		s.logger.Error("state.initialize-cluster.error", err)
		return nil, err
	}

	return cluster, nil
}
