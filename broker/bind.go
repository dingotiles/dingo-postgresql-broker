package broker

import (
	"fmt"
	"strconv"

	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
)

// Bind returns access credentials for a service instance
func (bkr *Broker) Bind(instanceID string, bindingID string, details brokerapi.BindDetails) (brokerapi.BindingResponse, error) {
	cluster := serviceinstance.NewCluster(instanceID, brokerapi.ProvisionDetails{}, bkr.EtcdClient, bkr.Config, bkr.Logger)

	key := fmt.Sprintf("/routing/allocation/%s", cluster.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, false)
	if err != nil {
		bkr.Logger.Error("bind.routing-allocation.get", err)
		return brokerapi.BindingResponse{}, fmt.Errorf("Internal error: no published port for provisioned cluster")
	}
	publicPort, err := strconv.ParseInt(resp.Node.Value, 10, 64)
	if err != nil {
		bkr.Logger.Error("bind.routing-allocation.parse-int", err)
		return brokerapi.BindingResponse{}, fmt.Errorf("Internal error: published port is not an integer (%s)", resp.Node.Value)
	}

	routerHost := bkr.Config.Router.Hostname
	username := "replicator"
	password := "replicator"
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres", username, password, routerHost, publicPort)
	jdbc := fmt.Sprintf("jdbc:postgresql://%s:%d/postgres?username=%s&password=%s", routerHost, publicPort, username, password)
	return brokerapi.BindingResponse{
		Credentials: brokerapi.CredentialsHash{
			Host:     routerHost,
			Port:     publicPort,
			Username: username,
			Password: password,
			URI:      uri,
			JDBCURI:  jdbc,
		},
	}, nil
}
