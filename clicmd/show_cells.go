package clicmd

import (
	"fmt"
	"log"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/codegangsta/cli"
)

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
