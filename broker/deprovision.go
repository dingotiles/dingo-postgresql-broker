package broker

import (
	"fmt"

	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
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

	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Logger)
	err = cluster.Load()
	if err != nil {
		return false, err
	}

	clusterRequest := servicechange.NewRequest(cluster, 0, 20)
	clusterRequest.Perform()

	// TODO cleanup KV
	bkr.EtcdClient.Delete(fmt.Sprintf("/service/%s", instanceID), true)
	bkr.EtcdClient.Delete(fmt.Sprintf("/serviceinstances/%s", instanceID), true)
	bkr.EtcdClient.Delete(fmt.Sprintf("/routing/allocation/%s", instanceID), true)

	return false, nil
}
