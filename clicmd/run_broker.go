package clicmd

import (
	"github.com/codegangsta/cli"
	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker"
)

// RunBroker runs the Cloud Foundry service broker API
func RunBroker(c *cli.Context) {
	cfg := loadConfig(c.String("config"))

	etcdClient := backend.NewEtcdClient(cfg.KVStore.Machines, "/")

	broker := broker.NewBroker(etcdClient, cfg)
	broker.Run()
}
