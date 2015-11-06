package broker

import (
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

	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Logger)
	err = cluster.Load()
	if err != nil {
		return false, err
	}

	clusterRequest := servicechange.NewRequest(cluster, 0, 20)
	clusterRequest.Perform()
	return false, nil
}
