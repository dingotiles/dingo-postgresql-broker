package broker

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Provision a new service instance
func (broker *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	// if missing, create /clusters/instanceID; else error/redirect to .Update()?
	cluster := serviceinstance.NewCluster(instanceID)
	clusterRequest := servicechange.NewRequest(cluster)
	clusterRequest.Perform()
	return brokerapi.ProvisioningResponse{}, true, nil
}
