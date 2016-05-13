package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/pivotal-golang/lager"
)

type State interface {

	// ClusterExists returns true if cluster already exists
	ClusterExists(clusterID string) bool
	// InitializeCluster(clusterID string) Cluster
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

// func (s *etcdStater) InitializeCluster(instanceID, details brokerapi.ProvisionDetails) Cluster {
// }
