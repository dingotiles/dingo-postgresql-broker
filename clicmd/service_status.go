package clicmd

import (
	"encoding/json"
	"fmt"
	"log"
	"path"

	"github.com/dingotiles/patroni-broker/backend"
	"github.com/dingotiles/patroni-broker/patroni"
	"github.com/codegangsta/cli"
)

// ServiceStatus displays to the terminal the status of all service clusters
func ServiceStatus(c *cli.Context) {
	cfg := loadConfig(c.String("config"))
	etcdClient := backend.NewEtcdClient(cfg.KVStore.Machines, "/")
	patroniServices, err := etcdClient.Get("/service", true, true)
	if err != nil {
		log.Fatal(err)
	}

	for _, serviceCluster := range patroniServices.Node.Nodes {
		serviceID := path.Base(serviceCluster.Key)
		for _, serviceData := range serviceCluster.Nodes {
			// find /service/ID/members list
			membersKey := fmt.Sprintf("%s/members", serviceCluster.Key)
			if serviceData.Key == membersKey {
				for _, member := range serviceData.Nodes {
					memberData := patroni.ServiceMemberData{}
					err := json.Unmarshal([]byte(member.Value), &memberData)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Printf("%s %s %s %s\n", serviceID, memberData.HostPort, memberData.Role, memberData.State)
				}
			}
		}
	}
}
