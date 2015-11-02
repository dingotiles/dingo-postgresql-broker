package main

import (
	"math/rand"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/broker"
	"github.com/codegangsta/cli"
)

func runBroker(c *cli.Context) {
	broker := broker.NewBroker()

	broker.Run()
}

func main() {
	rand.Seed(4200)

	app := cli.NewApp()
	app.Name = "patroni-broker"
	app.Version = "0.1.0"
	app.Usage = "Cloud Foundry service broker to run Patroni clusters"
	app.Commands = []cli.Command{
		{
			Name:   "broker",
			Usage:  "run the broker",
			Flags:  []cli.Flag{},
			Action: runBroker,
		},
	}
	app.Run(os.Args)

}
