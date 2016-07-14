package state

import (
	"encoding/json"
	"fmt"
	"regexp"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/pivotal-golang/lager"
)

const (
	LeaderRole  = "LeaderRole"
	ReplicaRole = "ReplicaRole"
)

type StateEtcd struct {
	etcdApi etcd.KeysAPI
	prefix  string
	logger  lager.Logger
	patroni *patroni.Patroni
}

type State interface {
	ClusterExists(structs.ClusterID) bool
	SaveCluster(structs.ClusterState) error
	LoadCluster(structs.ClusterID) (structs.ClusterState, error)
	LoadAllClusters() ([]*structs.ClusterState, error)
	DeleteCluster(structs.ClusterID) error
}

func NewStateEtcd(etcdConfig config.Etcd, logger lager.Logger) (*StateEtcd, error) {
	return NewStateEtcdWithPrefix(etcdConfig, "", logger)
}

func NewStateEtcdWithPrefix(etcdConfig config.Etcd, prefix string, logger lager.Logger) (*StateEtcd, error) {
	state := &StateEtcd{
		prefix: prefix,
		logger: logger,
	}

	patroniClient, err := patroni.NewPatroni(etcdConfig, logger)
	if err != nil {
		return nil, err
	}
	state.patroni = patroniClient

	state.etcdApi, err = state.setupEtcd(etcdConfig)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *StateEtcd) SaveCluster(clusterState structs.ClusterState) error {
	s.logger.Info("cluster-state.save", lager.Data{
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

func (s *StateEtcd) setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}

func (s *StateEtcd) ClusterExists(instanceID structs.ClusterID) bool {
	ctx := context.Background()
	s.logger.Info("state.cluster-exists")
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, instanceID)
	_, err := s.etcdApi.Get(ctx, key, &etcd.GetOptions{})
	return err == nil
}

// LoadCluster fetches the latest data/state for specific cluster
func (s *StateEtcd) LoadCluster(instanceID structs.ClusterID) (cluster structs.ClusterState, err error) {
	ctx := context.Background()
	s.logger.Info("state.load-cluster-state")

	key := fmt.Sprintf("%s/service/%s/state", s.prefix, instanceID)
	resp, err := s.etcdApi.Get(ctx, key, &etcd.GetOptions{})
	if err != nil {
		s.logger.Error("state.load-cluster-state.error", err)
		return
	}
	json.Unmarshal([]byte(resp.Node.Value), &cluster)

	leaderID, _ := s.patroni.ClusterLeader(instanceID)

	for _, node := range cluster.Nodes {
		if node.ID == leaderID {
			node.Role = LeaderRole
		} else {
			node.Role = ReplicaRole
		}
	}
	return
}

func (s *StateEtcd) LoadAllClusters() (clusters []*structs.ClusterState, err error) {
	ctx := context.Background()
	servicesKey := fmt.Sprintf("%s/service", s.prefix)
	services, err := s.etcdApi.Get(ctx, servicesKey, &etcd.GetOptions{Recursive: false})
	if err != nil {
		return
	}

	instanceIDRegExp, _ := regexp.Compile("/service/(.*)")
	for _, service := range services.Node.Nodes {
		instanceID := instanceIDRegExp.FindStringSubmatch(service.Key)[1]
		cluster, err := s.LoadCluster(structs.ClusterID(instanceID))
		if err != nil {
			return clusters, err
		}
		clusters = append(clusters, &cluster)
	}
	return
}

func (s *StateEtcd) DeleteCluster(instanceID structs.ClusterID) error {
	ctx := context.Background()
	s.logger.Info("state.delete-cluster-state")
	key := fmt.Sprintf("%s/service/%s", s.prefix, instanceID)

	_, err := s.etcdApi.Delete(ctx, key, &etcd.DeleteOptions{Recursive: true})
	if err != nil {
		s.logger.Error("state.delete-cluster-state", err)
	}

	return err
}
