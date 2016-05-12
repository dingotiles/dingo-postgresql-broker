package state

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/pivotal-golang/lager"
)

type State interface {

	// ClusterExists returns true if cluster already exists
	ClusterExists(clusterId string) bool
	// InitializeCluster(clusterId string) Cluster
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

func (s *etcdState) ClusterExists(instanceId string) bool {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", instanceId)
	_, err := s.etcd.Get(key, false, true)
	return err == nil
}

// func (s *etcdStater) InitializeCluster(instanceId, details brokerapi.ProvisionDetails) Cluster {
// }
