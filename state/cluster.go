package state

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/pivotal-golang/lager"
)

// Cluster describes a real/proposed cluster of nodes
type Cluster struct {
	etcdClient backend.EtcdClient
	logger     lager.Logger
	meta       structs.ClusterData
}

func (c *Cluster) MetaData() structs.ClusterData {
	return c.meta
}

func (c *Cluster) SetTargetNodeCount(count int) error {
	c.restoreState()
	c.meta.TargetNodeCount = count
	err := c.writeState()
	if err != nil {
		c.logger.Error("cluster.set-target-node-count.error", err)
		return err
	}
	return nil
}

func (c *Cluster) PortAllocation() (int64, error) {
	key := fmt.Sprintf("/routing/allocation/%s", c.meta.InstanceID)
	resp, err := c.etcdClient.Get(key, false, false)
	if err != nil {
		c.logger.Error("cluster.routing-allocation.get", err)
		return 0, err
	}
	publicPort, err := strconv.ParseInt(resp.Node.Value, 10, 64)
	if err != nil {
		c.logger.Error("cluster.routing-allocation.parse-int", err)
		return 0, err
	}
	return publicPort, nil
}

// WaitForRoutingPortAllocation blocks until the routing tier has allocated a public port
func (cluster *Cluster) WaitForRoutingPortAllocation() (err error) {
	for index := 0; index < 10; index++ {
		key := fmt.Sprintf("/routing/allocation/%s", cluster.MetaData().InstanceID)
		resp, err := cluster.etcdClient.Get(key, false, false)
		if err != nil {
			cluster.logger.Debug("cluster.wait-for-port", lager.Data{})
		} else {
			cluster.meta.AllocatedPort = resp.Node.Value
			cluster.logger.Info("cluster.wait-for-port", lager.Data{"allocated_port": cluster.meta.AllocatedPort})
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	cluster.logger.Error("cluster.wait-for-port.timeout", err, lager.Data{"err": err})
	return err
}

func (c *Cluster) writeState() error {
	c.logger.Info("cluster.write-state", lager.Data{"meta": c.meta})
	key := fmt.Sprintf("/serviceinstances/%s/plan_id", c.meta.InstanceID)
	_, err := c.etcdClient.Set(key, c.meta.PlanID, 0)
	if err != nil {
		c.logger.Error("cluster.write-state.error", err)
		return err
	}
	key = fmt.Sprintf("/serviceinstances/%s/meta", c.meta.InstanceID)
	_, err = c.etcdClient.Set(key, c.meta.Json(), 0)
	if err != nil {
		c.logger.Error("cluster.write-state.error", err)
		return err
	}
	return nil
}

func (c *Cluster) restoreState() error {
	c.logger.Info("cluster.restore-state")
	key := fmt.Sprintf("/serviceinstances/%s/meta", c.meta.InstanceID)
	resp, err := c.etcdClient.Get(key, false, false)
	if err != nil {
		c.logger.Error("restore-state.error", err)
		return err
	}
	c.meta = *structs.ClusterDataFromJson(resp.Node.Value)
	return nil
}

func (c *Cluster) deleteState() error {
	var lastError error
	resp, err := c.etcdClient.Delete(fmt.Sprintf("/serviceinstances/%s", c.meta.InstanceID), true)
	if err != nil {
		c.logger.Error("cluster.delete-state.err", err, lager.Data{"etcd-response": resp})
		lastError = err
	}
	resp, err = c.etcdClient.Delete(fmt.Sprintf("/routing/allocation/%s", c.meta.InstanceID), true)
	if err != nil {
		c.logger.Error("cluster.delete-allocation.err", err, lager.Data{"etcd-response": resp})
		lastError = err
	}

	// clear out etcd data that would eventually timeout; to allow immediate recreation if required by user
	resp, err = c.etcdClient.Delete(fmt.Sprintf("/service/%s/members", c.meta.InstanceID), true)
	if err != nil {
		c.logger.Error("cluster.delete-members.err", err, lager.Data{"etcd-response": resp})
		lastError = err
	}
	resp, err = c.etcdClient.Delete(fmt.Sprintf("/service/%s/optime", c.meta.InstanceID), true)
	if err != nil {
		c.logger.Error("cluster.delete-optime.err", err, lager.Data{"etcd-response": resp})
		lastError = err
	}
	resp, err = c.etcdClient.Delete(fmt.Sprintf("/service/%s/leader", c.meta.InstanceID), true)
	if err != nil {
		c.logger.Error("cluster.delete-leader.err", err, lager.Data{"etcd-response": resp})
		lastError = err
	}
	return lastError
}
