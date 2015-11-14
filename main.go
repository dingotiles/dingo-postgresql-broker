package main

import (
	"math/rand"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/clicmd"
	"github.com/codegangsta/cli"
)

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
			Action: clicmd.RunBroker,
		},
		{
			Name:  "service-status",
			Usage: "status of all service clusters",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Value: "config.yml",
					Usage: "path to YAML config file",
				},
			},
			Action: clicmd.ServiceStatus,
		},
	}
	app.Run(os.Args)
}
