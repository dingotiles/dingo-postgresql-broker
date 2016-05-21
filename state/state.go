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
	SaveCluster(cluster structs.ClusterState) error
	LoadClusterState(instanceID string) (structs.ClusterState, error)
	DeleteClusterState(instanceID string) error
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
		s.logger.Error("save-cluster.set-plan-id.error", err)
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
	key := fmt.Sprintf("/serviceinstances/%s/nodes", instanceID)
	_, err := s.etcd.Get(key, false, true)
	return err == nil
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

func (s *etcdState) DeleteClusterState(instanceID string) error {
	ctx := context.TODO()
	s.logger.Info("state.delete-cluster-state")
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, instanceID)
	planKey := fmt.Sprintf("%s/service/%s/plan_id", s.prefix, instanceID)

	var lastError error
	_, err := s.etcdApi.Delete(ctx, key, &etcd.DeleteOptions{})
	if err != nil {
		s.logger.Error("state.delete-cluster-state", err)
		lastError = err
	}
	_, err = s.etcdApi.Delete(ctx, planKey, &etcd.DeleteOptions{})
	if err != nil {
		s.logger.Error("state.delete-cluster-state", err)
		lastError = err
	}

	err = s.deletePatroniState(ctx, instanceID)
	if err != nil {
		s.logger.Error("cluster.delete-patroni-state", err)
	}

	return lastError
}

func (s *etcdState) deletePatroniState(ctx context.Context, instanceID string) error {
	var lastError, err error
	// clear out etcd data that would eventually timeout; to allow immediate recreation if required by user
	_, err = s.etcdApi.Delete(ctx, fmt.Sprintf("%s/service/%s/members", s.prefix, instanceID), &etcd.DeleteOptions{})
	if err != nil {
		lastError = err
	}
	_, err = s.etcdApi.Delete(ctx, fmt.Sprintf("%s/service/%s/optime", s.prefix, instanceID), &etcd.DeleteOptions{})
	if err != nil {
		lastError = err
	}
	_, err = s.etcdApi.Delete(ctx, fmt.Sprintf("%s/service/%s/leader", s.prefix, instanceID), &etcd.DeleteOptions{})
	if err != nil {
		lastError = err
	}
	return lastError
}
