package broker

import (
	"fmt"

	"github.com/coreos/go-etcd/etcd"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
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

	cluster := state.NewClusterFromProvisionDetails(instanceID, details, bkr.etcdClient, bkr.config, bkr.logger)
	logger := bkr.logger
	err = cluster.Load()
	if err != nil {
		logger.Error("load", err)
		return false, err
	}

	clusterRequest := bkr.scheduler.NewRequest(cluster, 0)
	bkr.scheduler.Execute(clusterRequest)

	var resp *etcd.Response
	resp, err = bkr.etcdClient.Delete(fmt.Sprintf("/serviceinstances/%s", instanceID), true)
	if err != nil {
		logger.Error("etcd-delete.serviceinstances.err", err, lager.Data{"response": resp})
	}
	resp, err = bkr.etcdClient.Delete(fmt.Sprintf("/routing/allocation/%s", instanceID), true)
	if err != nil {
		logger.Error("etcd-delete.routing-allocation.err", err, lager.Data{"response": resp})
	}

	// clear out etcd data that would eventually timeout; to allow immediate recreation if required by user
	resp, err = bkr.etcdClient.Delete(fmt.Sprintf("/service/%s/members", instanceID), true)
	if err != nil {
		logger.Error("etcd-delete.service-members.err", err, lager.Data{"response": resp})
	}
	resp, err = bkr.etcdClient.Delete(fmt.Sprintf("/service/%s/optime", instanceID), true)
	if err != nil {
		logger.Error("etcd-delete.service-optime.err", err, lager.Data{"response": resp})
	}
	resp, err = bkr.etcdClient.Delete(fmt.Sprintf("/service/%s/leader", instanceID), true)
	if err != nil {
		logger.Error("etcd-delete.service-leader.err", err, lager.Data{"response": resp})
	}
	logger.Info("etcd-delete.done")
	return false, nil
}
