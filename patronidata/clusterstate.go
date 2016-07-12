package patronidata

import "github.com/dingotiles/dingo-postgresql-broker/broker/structs"

// ClusterDataWrapper allows access to cluster state for a specific cluster
type ClusterDataWrapperReal struct {
	patroni    *Patroni
	instanceID structs.ClusterID
}

type ClusterDataWrapper interface {
	LoadCluster() (clusterState structs.ClusterState, err error)
	WaitTilMemberRunning(memberID string) error
}

// NewClusterDataWrapper creates a ClusterDataWrapper
func NewClusterDataWrapper(patroni *Patroni, instanceID structs.ClusterID) ClusterDataWrapper {
	return ClusterDataWrapperReal{
		patroni:    patroni,
		instanceID: instanceID,
	}
}

// LoadCluster fetches the latest data/state for specific cluster
func (cluster ClusterDataWrapperReal) LoadCluster() (clusterState structs.ClusterState, err error) {
	return
}

func (cluster ClusterDataWrapperReal) WaitTilMemberRunning(memberID string) error {
	for {
		_, err := cluster.LoadCluster()
		if err != nil {
			return err
		}
	}
	return nil
}
