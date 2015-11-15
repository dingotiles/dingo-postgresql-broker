package clicmd

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/config"
	"github.com/codegangsta/cli"
)

func loadConfig(configPath string) (cfg *config.Config) {
	if os.Getenv("PATRONI_BROKER_CONFIG") != "" {
		configPath = os.Getenv("PATRONI_BROKER_CONFIG")
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// ShowCells displays the status of each backend cell/server
func ShowCells(c *cli.Context) {
	cfg := loadConfig(c.String("config"))
	etcdClient := backend.NewEtcdClient(cfg.KVStore.Machines, "/")
	serviceinstances, err := etcdClient.Get("/serviceinstances", true, true)
	if err != nil {
		log.Fatal(err)
	}

	cellsContainerCounter := map[string]int{}
	// cellsNodes := map[string][]string
	for _, serviceinstance := range serviceinstances.Node.Nodes {
		serviceNodesResp, err := etcdClient.Get(fmt.Sprintf("%s/nodes", serviceinstance.Key), false, false)
		if err != nil {
			log.Fatal(err)
		}
		for _, serviceNode := range serviceNodesResp.Node.Nodes {
			backendResp, err := etcdClient.Get(fmt.Sprintf("%s/backend", serviceNode.Key), false, false)
			if err != nil {
				log.Fatal(err)
			}
			backendGUID := backendResp.Node.Value
			cellsContainerCounter[backendGUID]++
		}
	}

	for backendGUID, containerCount := range cellsContainerCounter {
		fmt.Printf("%d %s\n", containerCount, backendGUID)
	}
}
