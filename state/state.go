package state

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type State interface {

	// ClusterExists returns true if cluster already exists
	ClusterExists(instanceID string) bool
	SaveCluster(cluster structs.ClusterState) error
	LoadCluster(instanceID string) (structs.ClusterState, error)
	DeleteCluster(instanceID string) error
}

type etcdState struct {
	etcdApi etcd.KeysAPI
	prefix  string
	logger  lager.Logger
}

func NewState(etcdConfig config.Etcd, logger lager.Logger) (State, error) {
	state := &etcdState{
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

	ctx := context.Background()
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
	ctx := context.Background()
	s.logger.Info("state.cluster-exists")
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, instanceID)
	_, err := s.etcdApi.Get(ctx, key, &etcd.GetOptions{})
	return err == nil
}
func (s *etcdState) LoadCluster(instanceID string) (structs.ClusterState, error) {
	var cluster structs.ClusterState
	ctx := context.Background()
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

func (s *etcdState) DeleteCluster(instanceID string) error {
	ctx := context.Background()
	s.logger.Info("state.delete-cluster-state")
	key := fmt.Sprintf("%s/service/%s", s.prefix, instanceID)

	_, err := s.etcdApi.Delete(ctx, key, &etcd.DeleteOptions{Recursive: true})
	if err != nil {
		s.logger.Error("state.delete-cluster-state", err)
	}

	return err
}
