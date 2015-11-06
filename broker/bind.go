package broker

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Bind returns access credentials for a service instance
func (bkr *Broker) Bind(instanceID string, bindingID string, details brokerapi.BindDetails) (brokerapi.BindingResponse, error) {
	cluster := serviceinstance.NewCluster(instanceID, brokerapi.ProvisionDetails{}, bkr.EtcdClient, bkr.Logger)

	key := fmt.Sprintf("/routing/allocation/%s", cluster.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, false)
	if err != nil {
		return brokerapi.BindingResponse{}, fmt.Errorf("Internal error: no published port for provisioned cluster")
	}
	publicPort, err := strconv.ParseInt(resp.Node.Value, 10, 64)
	if err != nil {
		return brokerapi.BindingResponse{}, fmt.Errorf("Internal error: published port is not an integer (%s)", resp.Node.Value)
	}

	return brokerapi.BindingResponse{
		Credentials: brokerapi.CredentialsHash{
			Host: "10.10.10.10",
			Port: publicPort,
		},
	}, nil
}
