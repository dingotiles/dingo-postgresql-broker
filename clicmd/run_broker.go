package clicmd

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker"
)

// RunBroker runs the Cloud Foundry service broker API
func RunBroker(c *cli.Context) {
	cfg := loadConfig(c.String("config"))

	etcdClient := backend.NewEtcdClient(cfg.Etcd.Machines, "/")

	broker, err := broker.NewBroker(etcdClient, cfg)
	if err != nil {
		fmt.Println("Could not start broker")
		os.Exit(1)
		return
	}

	broker.Run()
}
