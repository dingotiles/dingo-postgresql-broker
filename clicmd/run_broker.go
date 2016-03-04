package clicmd

import (
	"github.com/dingotiles/patroni-broker/backend"
	"github.com/dingotiles/patroni-broker/broker"
	"github.com/codegangsta/cli"
)

// RunBroker runs the Cloud Foundry service broker API
func RunBroker(c *cli.Context) {
	cfg := loadConfig(c.String("config"))
	etcdClient := backend.NewEtcdClient(cfg.KVStore.Machines, "/")

	broker := broker.NewBroker(etcdClient, cfg)
	broker.Run()
}
