package state

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"

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
	LoadClusterState(instanceID string) (structs.ClusterState, error)
}

type etcdState struct {
	etcd    backend.EtcdClient
	etcdApi etcd.KeysAPI
	prefix  string
	logger  lager.Logger
}

func NewState(etcdConfig config.Etcd, etcdClient backend.EtcdClient, logger lager.Logger) (State, error) {
	state := &etcdState{
		etcd:   etcdClient,
		logger: logger,
		prefix: "",
	}

	var err error
	state.etcdApi, err = state.setupEtcd(etcdConfig)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func NewStateWithPrefix(etcdConfig config.Etcd, prefix string, logger lager.Logger) (State, error) {
	state := &etcdState{
		prefix: prefix,
		etcd:   backend.NewEtcdClient(etcdConfig.Machines, prefix),
		logger: logger,
	}

	var err error
	state.etcdApi, err = state.setupEtcd(etcdConfig)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *etcdState) SaveCluster(clusterState structs.ClusterState) error {
	s.logger.Info("save-clusterState", lager.Data{
		"cluster": clusterState,
	})

	ctx := context.TODO()
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, clusterState.InstanceID)

	data, err := json.Marshal(clusterState)
	if err != nil {
		s.logger.Error("save-cluster.marshal", err)
	}

	_, err = s.etcdApi.Set(ctx, key, string(data), &etcd.SetOptions{})
	if err != nil {
		s.logger.Error("save-cluster.set", err)
		return err
	}

	planKey := fmt.Sprintf("%s/service/%s/plan_id", s.prefix, clusterState.InstanceID)
	_, err = s.etcdApi.Set(ctx, planKey, clusterState.PlanID, &etcd.SetOptions{})
	if err != nil {
		s.logger.Error("save-cluster.set-plan-id", err)
		return err
	}

	cluster := &Cluster{
		etcdClient: s.etcd,
		logger:     s.logger,
		meta:       clusterState.MetaData(),
	}
	err = cluster.writeState()
	if err != nil {
		s.logger.Error("save-cluster.write-state", err)
		return err
	}
	for _, n := range clusterState.Nodes() {
		cluster.AddNode(*n)
	}

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

func (s *etcdState) LoadClusterState(instanceID string) (structs.ClusterState, error) {
	var cluster structs.ClusterState
	ctx := context.TODO()
	s.logger.Info("state.load-cluster-state")
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, instanceID)

	resp, err := s.etcdApi.Get(ctx, key, &etcd.GetOptions{})
	if err != nil {
		s.logger.Error("state.load-cluster-state.error", err)
		return cluster, err
	}
	json.Unmarshal([]byte(resp.Node.Value), &cluster)
	return cluster, nil
}
