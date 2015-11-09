package broker

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)
	clusterRequest := servicechange.NewRequest(cluster, 2, 20)
	clusterRequest.Perform()
	cluster.WaitForRoutingPortAllocation()
	return brokerapi.ProvisioningResponse{}, false, nil
}
