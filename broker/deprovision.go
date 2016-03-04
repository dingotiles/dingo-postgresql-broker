package broker

import (
	"fmt"

	"github.com/dingotiles/patroni-broker/servicechange"
	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Deprovision service instance
func (bkr *Broker) Deprovision(instanceID string, deprovDetails brokerapi.DeprovisionDetails, acceptsIncomplete bool) (async bool, err error) {
	details := brokerapi.ProvisionDetails{
		ServiceID: deprovDetails.ServiceID,
		PlanID:    deprovDetails.PlanID,
	}
	if details.ServiceID == "" || details.PlanID == "" {
		return false, fmt.Errorf("API error - provide service_id and plan_id as URL parameters")
	}

	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)
	err = cluster.Load()
	if err != nil {
		return false, err
	}

	clusterRequest := servicechange.NewRequest(cluster, 0, 20)
	clusterRequest.Perform()

	bkr.EtcdClient.Delete(fmt.Sprintf("/serviceinstances/%s", instanceID), true)
	bkr.EtcdClient.Delete(fmt.Sprintf("/routing/allocation/%s", instanceID), true)

	// cluster technically not deleted until until /service/%s/members is empty
	// members, err := bkr.EtcdClient.Get(fmt.Sprintf("/service/%s/members", instanceID), false, false)
	// counter := 1
	// for err == nil && members.Node.Nodes != nil {
	// 	fmt.Printf("%ds waiting for %d nodes to be removed from %s\n", counter, len(members.Node.Nodes), fmt.Sprintf("/service/%s/members", instanceID))
	// 	counter++
	// 	time.Sleep(1 * time.Second)
	// 	members, err = bkr.EtcdClient.Get(fmt.Sprintf("/service/%s/members", instanceID), false, false)
	// }

	return false, nil
}
