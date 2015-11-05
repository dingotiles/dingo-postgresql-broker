package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/broker"
	"github.com/codegangsta/cli"
)

func runBroker(c *cli.Context) {
	machines := []string{"http://127.0.0.1:2379"}
	etcdClient := backend.NewEtcdClient(machines, "/playtime")

	broker := broker.NewBroker(etcdClient)
	broker.Run()
}

func runDevSilliness(c *cli.Context) {
	machines := []string{"http://127.0.0.1:2379"}
	etcdClient := backend.NewEtcdClient(machines, "/playtime")
	backendBkr := backend.Backend{GUID: "5ac91960-0cfa-4c31-90ab-3f6442ac637d", URI: "http://10.244.21.6", Username: "containers", Password: "containers"}
	err := backend.AddBackendToEtcd(etcdClient, backendBkr)
	if err != nil {
		log.Fatal(err)
	}
	backends, err := backend.LoadBackendsFromEtcd(etcdClient)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", backends)
}

func main() {
	rand.Seed(5000)

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
