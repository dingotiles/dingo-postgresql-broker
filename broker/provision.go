package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/dingotiles/patroni-broker/servicechange"
	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	cluster := serviceinstance.NewCluster(instanceID, details, bkr.EtcdClient, bkr.Config, bkr.Logger)

	if details.ServiceID == "" && details.PlanID == "" {
		return bkr.Recreate(instanceID, acceptsIncomplete)
	}

	logger := cluster.Logger
	logger.Info("provision.start", lager.Data{})

	if cluster.Exists() {
		return resp, false, fmt.Errorf("service instance %s already exists", instanceID)
	}

	// 1-node default cluster
	nodeCount := 1
	nodeSize := 20 // meaningless at moment
	if details.Parameters["node-count"] != nil {
		rawNodeCount := details.Parameters["node-count"]
		nodeCount = int(rawNodeCount.(float64))
	}
	if nodeCount < 1 {
		logger.Info("provision.start.node-count-too-low", lager.Data{"node-count": nodeCount})
		nodeCount = 1
	}
	clusterRequest := servicechange.NewRequest(cluster, int(nodeCount), nodeSize)

	err = clusterRequest.Perform()
	if err != nil {
		logger.Error("provision.perform.error", err)
		return resp, false, err
	}

	err = cluster.WaitForAllRunning()
	if err == nil {
		// if cluster is running, then wait until routing port operational
		err = cluster.WaitForRoutingPortAllocation()
	}

	if err != nil {
		logger.Error("provision.running.error", err)
	} else {
		logger.Info("provision.running.success", lager.Data{"cluster": cluster.ClusterData()})
		bkr.triggerClusterDataBackup(cluster)
	}
	return resp, false, err
}

func (bkr *Broker) triggerClusterDataBackup(cluster *serviceinstance.Cluster) {
	logger := cluster.Logger
	callback := bkr.Config.Callbacks.ClusterDataBackup
	if callback == nil {
		logger.Info("provision.success.callback.noop")
		return
	}

	data, err := json.Marshal(cluster.ClusterData())
	if err != nil {
		logger.Error("provision.success.callback.data-marshal", err)
		return
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("provision.success.callback.stdin-pipe", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("provision.success.callback.stdout-pipe", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("provision.success.callback.stderr-pipe", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("provision.success.callback.start", err)
		return
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		defer stdin.Close()
		io.Copy(stdin, bytes.NewBufferString(string(data)))
	}()
	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		defer wg.Done()
		io.Copy(os.Stderr, stderr)
	}()
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		logger.Error("provision.success.callback.cmd", err)
		return
	}
	logger.Info("provision.success.callback.done")
}
