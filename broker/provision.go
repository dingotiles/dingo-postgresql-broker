package broker

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	// if missing, create /clusters/instanceID; else error/redirect to .Update()?
	cluster := serviceinstance.NewCluster(instanceID)
	clusterRequest := servicechange.NewRequest(cluster, 2, 20)
	clusterRequest.Perform(bkr.Logger)
	return brokerapi.ProvisioningResponse{DashboardURL: "foo"}, true, nil
}
