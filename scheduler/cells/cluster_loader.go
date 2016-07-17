package cells

import "github.com/dingotiles/dingo-postgresql-broker/broker/structs"

type ClusterLoader interface {
	LoadAllClusters() ([]*structs.ClusterState, error)
}
