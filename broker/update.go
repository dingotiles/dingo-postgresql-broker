package broker

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Update service instance
func (bkr *Broker) Update(instanceID string, details brokerapi.UpdateDetails, acceptsIncomplete bool) (bool, error) {
	provisionDetails := brokerapi.ProvisionDetails{
		ServiceID:  details.ServiceID,
		PlanID:     details.PlanID,
		Parameters: details.Parameters,
	}
	cluster := serviceinstance.NewCluster(instanceID, provisionDetails, bkr.EtcdClient, bkr.Logger)
	cluster.Load()
	clusterRequest := servicechange.NewRequest(cluster, 4, 20)
	clusterRequest.Perform()
	return true, nil
}
