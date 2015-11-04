package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/broker"
	"github.com/codegangsta/cli"
)

func runBroker(c *cli.Context) {
	broker := broker.NewBroker()

	broker.Run()
}

func runDevSilliness(c *cli.Context) {
	machines := []string{"http://127.0.0.1:2379"}
	err := broker.AddBackendToEtcd(broker.Backend{GUID: "boom"}, machines, "/")
	if err != nil {
		log.Fatal(err)
	}
	backends, err := broker.LoadBackendsFromEtcd(machines, "/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", backends)
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
		{
			Name:   "dev",
			Usage:  "invoke something internal",
			Flags:  []cli.Flag{},
			Action: runDevSilliness,
		},
	}
	app.Run(os.Args)

}
