package state

import (
	"fmt"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type State interface {

	// ClusterExists returns true if cluster already exists
	ClusterExists(instanceID string) bool
	InitializeCluster(clusterData *structs.ClusterData) (*Cluster, error)
	LoadCluster(instanceID string) (*Cluster, error)
	DeleteCluster(cluster *Cluster) error
	SaveCluster(cluster structs.ClusterState) error
}

type etcdState struct {
	etcd    backend.EtcdClient
	etcdApi etcd.KeysAPI
	logger  lager.Logger
}

func NewState(etcdConfig config.Etcd, etcdClient backend.EtcdClient, logger lager.Logger) (State, error) {
	state := &etcdState{
		etcd:   etcdClient,
		logger: logger,
	}

	var err error
	state.etcdApi, err = state.setupEtcd(etcdConfig)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *etcdState) SaveCluster(cluster structs.ClusterState) error {
	return nil
}

func (s *etcdState) setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
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

func (s *etcdState) DeleteCluster(cluster *Cluster) error {
	if err := cluster.deleteState(); err != nil {
		s.logger.Error("state.delete-cluster.error", err, lager.Data{"cluster": cluster.MetaData()})
		return err
	}
	return nil
}
