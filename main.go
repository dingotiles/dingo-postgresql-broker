package main

import (
	"log"
	"math/rand"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/broker"
	"github.com/cloudfoundry-community/patroni-broker/config"
	"github.com/codegangsta/cli"
)

func runBroker(c *cli.Context) {
	configPath := c.String("config")
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	etcdClient := backend.NewEtcdClient(config.KVStore.Machines, "/")

	broker := broker.NewBroker(etcdClient, config)
	broker.Run()
}

func main() {
	rand.Seed(5000)

	app := cli.NewApp()
	app.Name = "patroni-broker"
	app.Version = "0.1.0"
	app.Usage = "Cloud Foundry service broker to run Patroni clusters"
	app.Commands = []cli.Command{
		{
			Name:  "broker",
			Usage: "run the broker",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Value: "config.yml",
					Usage: "path to YAML config file",
				},
			},
			Action: runBroker,
		},
	}
	app.Run(os.Args)
}
