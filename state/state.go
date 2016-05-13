package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/pivotal-golang/lager"
)

type State interface {

	// ClusterExists returns true if cluster already exists
	ClusterExists(instanceID string) bool
	InitializeCluster(clusterData *structs.ClusterData) (*Cluster, error)
	LoadCluster(instanceID string) (*Cluster, error)
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
		logger:     s.logger,
		meta:       *clusterData,
	}
	err := cluster.writeState()
	if err != nil {
		s.logger.Error("state.initialize-cluster.error", err)
		return nil, err
	}

	return cluster, nil
}

func (s *etcdState) LoadCluster(instanceID string) (*Cluster, error) {
	cluster := &Cluster{
		etcdClient: s.etcd,
		logger: s.logger.Session("cluster", lager.Data{
			"instance-id": instanceID,
		}),
		meta: structs.ClusterData{InstanceID: instanceID},
	}
	err := cluster.restoreState()
	if err != nil {
		s.logger.Error("state.load-cluster.error", err)
		return nil, err
	}
	return cluster, nil
}
