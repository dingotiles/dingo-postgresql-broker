package cells

import (
	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

type CellsEtcd struct {
	etcdConfig config.Etcd
	etcdApi    etcd.KeysAPI
	prefix     string
	logger     lager.Logger
}

type Cells interface {
	LoadStatus() (*CellsHealth, error)
}

type CellsHealth map[string]*CellHealth

type CellHealth struct {
	NodeCount int
}

func NewCellsEtcd(etcdConfig config.Etcd, logger lager.Logger) (*CellsEtcd, error) {
	return NewCellsEtcdWithPrefix(etcdConfig, "", logger)
}

func NewCellsEtcdWithPrefix(etcdConfig config.Etcd, prefix string, logger lager.Logger) (*CellsEtcd, error) {
	cells := &CellsEtcd{
		etcdConfig: etcdConfig,
		prefix:     prefix,
		logger:     logger,
	}

	client, err := etcd.New(etcd.Config{Endpoints: etcdConfig.Machines})
	if err != nil {
		return nil, err
	}

	cells.etcdApi = etcd.NewKeysAPI(client)

	return cells, nil
}

// LoadStatus discovers the layout of cells that contain nodes
// Cells that do not contain nodes yet are not included in the results
func (cells *CellsEtcd) LoadStatus() (health *CellsHealth, err error) {
	state, err := state.NewStateEtcdWithPrefix(cells.etcdConfig, cells.prefix, cells.logger)
	if err != nil {
		return
	}
	clusters, err := state.LoadAllClusters()
	if err != nil {
		return
	}
	health = &CellsHealth{}
	for _, cluster := range clusters {
		for _, clusterNode := range cluster.Nodes {
			backendID := clusterNode.BackendID
			if (*health)[backendID] == nil {
				(*health)[backendID] = &CellHealth{}
			}
			(*health)[backendID].NodeCount += 1
		}
	}
	return
}
