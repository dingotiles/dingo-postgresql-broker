package state

import (
	"encoding/json"
	"fmt"
	"regexp"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type StateEtcd struct {
	etcdApi etcd.KeysAPI
	prefix  string
	logger  lager.Logger
}

func NewStateEtcd(etcdConfig config.Etcd, logger lager.Logger) (*StateEtcd, error) {
	return NewStateEtcdWithPrefix(etcdConfig, "", logger)
}

func NewStateEtcdWithPrefix(etcdConfig config.Etcd, prefix string, logger lager.Logger) (*StateEtcd, error) {
	state := &StateEtcd{
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

// SaveCluster stores the known state of the cluster in the /state endpoint
func (s *StateEtcd) SaveCluster(clusterState structs.ClusterState) (err error) {
	s.logger.Info("state.save-cluster", lager.Data{
		"cluster": clusterState,
	})

	ctx := context.Background()
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, clusterState.InstanceID)

	data, err := json.Marshal(clusterState)
	if err != nil {
		s.logger.Error("state.save-cluster.marshal", err, lager.Data{"instance-id": clusterState.InstanceID})
		return
	}

	_, err = s.etcdApi.Set(ctx, key, string(data), &etcd.SetOptions{})
	if err != nil {
		s.logger.Error("state.save-cluster.set", err, lager.Data{"instance-id": clusterState.InstanceID})
		return
	}

	return
}

func (s *StateEtcd) setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}

// ClusterExists looks up if this cluster has its state stored in etcd
func (s *StateEtcd) ClusterExists(instanceID structs.ClusterID) bool {
	ctx := context.Background()
	key := fmt.Sprintf("%s/service/%s/state", s.prefix, instanceID)
	_, err := s.etcdApi.Get(ctx, key, &etcd.GetOptions{})
	s.logger.Info("state.cluster-exists", lager.Data{"instance-id": instanceID, "exists": err == nil})
	return err == nil
}

// LoadCluster fetches the latest data/state for specific cluster
func (s *StateEtcd) LoadCluster(instanceID structs.ClusterID) (cluster structs.ClusterState, err error) {
	ctx := context.Background()
	s.logger.Info("state.load-cluster", lager.Data{"instance-id": instanceID})

	key := fmt.Sprintf("%s/service/%s", s.prefix, instanceID)
	resp, err := s.etcdApi.Get(ctx, key, &etcd.GetOptions{Recursive: true})
	if err != nil {
		s.logger.Error("state.load-cluster.error", err, lager.Data{"instance-id": instanceID})
		return
	}

	nodes := []*structs.Node{}
	for _, path := range resp.Node.Nodes {
		if match, _ := regexp.MatchString(fmt.Sprintf("%s/state", key), path.Key); match == true {
			json.Unmarshal([]byte(path.Value), &cluster)
		}
		if match, _ := regexp.MatchString(fmt.Sprintf("%s/nodes", key), path.Key); match == true {
			for _, member := range path.Nodes {
				var node structs.Node
				json.Unmarshal([]byte(member.Value), &node)
				if node.ID != "" && node.CellGUID != "" {
					nodes = append(nodes, &node)
				}
			}
		}
	}
	cluster.Nodes = nodes

	return
}

// LoadAllRunningClusters fetches the /state information for all running clusters
func (s *StateEtcd) LoadAllRunningClusters() (clusters []*structs.ClusterState, err error) {
	ctx := context.Background()
	servicesKey := fmt.Sprintf("%s/service", s.prefix)
	services, err := s.etcdApi.Get(ctx, servicesKey, &etcd.GetOptions{Recursive: false})
	if err != nil {
		return
	}

	instanceIDRegExp, _ := regexp.Compile("/service/(.*)")
	for _, service := range services.Node.Nodes {
		instanceID := instanceIDRegExp.FindStringSubmatch(service.Key)[1]
		cluster, _ := s.LoadCluster(structs.ClusterID(instanceID))
		clusters = append(clusters, &cluster)
	}
	return
}

func (s *StateEtcd) DeleteCluster(instanceID structs.ClusterID) error {
	ctx := context.Background()
	s.logger.Info("state.delete-cluster")
	key := fmt.Sprintf("%s/service/%s", s.prefix, instanceID)

	_, err := s.etcdApi.Delete(ctx, key, &etcd.DeleteOptions{Recursive: true})
	if err != nil {
		s.logger.Error("state.delete-cluster", err)
	}

	return err
}
