package clicmd

import (
	"log"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/broker"
	"github.com/cloudfoundry-community/patroni-broker/config"
	"github.com/codegangsta/cli"
)

// RunBroker runs the Cloud Foundry service broker API
func RunBroker(c *cli.Context) {
	configPath := c.String("config")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	etcdClient := backend.NewEtcdClient(cfg.KVStore.Machines, "/")

	broker := broker.NewBroker(etcdClient, cfg)
	broker.Run()
}
