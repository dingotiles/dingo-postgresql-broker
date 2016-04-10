package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/dingotiles/dingo-postgresql-broker/servicechange"
	"github.com/dingotiles/dingo-postgresql-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Recreate(instanceID string, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	logger := bkr.Logger.Session("recreate", lager.Data{
		"instance-id": instanceID,
	})

	logger.Info("start", lager.Data{})
	var clusterdata *serviceinstance.ClusterData
	err, clusterdata = bkr.restoreClusterDataBackup(instanceID)
	if err != nil {
		err = fmt.Errorf("Cannot recreate service from backup; unable to restore original service instance data: %s", err)
		return
	}

	cluster := serviceinstance.NewClusterFromRestoredData(instanceID, clusterdata, bkr.EtcdClient, bkr.Config, logger)
	logger = cluster.Logger
	logger.Info("me", lager.Data{"cluster": cluster})

	if cluster.Exists() {
		logger.Info("exists")
		err = fmt.Errorf("Service instance %s still exists in etcd, please clean it out before recreating cluster", instanceID)
		return
	} else {
		logger.Info("not-exists")
	}

	// Restore port allocation from cluster.Data
	key := fmt.Sprintf("/routing/allocation/%s", cluster.Data.InstanceID)
	_, err = cluster.EtcdClient.Set(key, cluster.Data.AllocatedPort, 0)
	if err != nil {
		logger.Error("routing-allocation.error", err)
		return
	}
	logger.Info("routing-allocation.restored", lager.Data{"allocated-port": cluster.Data.AllocatedPort})

	nodeSize := cluster.Data.NodeSize
	nodeCount := cluster.Data.NodeCount
	if nodeCount < 1 {
		nodeCount = 1
	}
	cluster.Data.NodeSize = 0
	cluster.Data.NodeCount = 0
	clusterRequest := servicechange.NewRequest(cluster, nodeCount, nodeSize)
	err = clusterRequest.Perform()
	if err != nil {
		logger.Error("provision.perform.error", err)
		return resp, false, err
	}

	// if port is allocated, then wait to confirm containers are running
	err = cluster.WaitForAllRunning()

	if err != nil {
		logger.Error("provision.running.error", err)
		return
	}
	logger.Info("provision.running.success", lager.Data{"cluster": cluster.ClusterData()})
	bkr.triggerClusterDataBackup(cluster)
	return
}

func (bkr *Broker) restoreClusterDataBackup(instanceID string) (err error, clusterdata *serviceinstance.ClusterData) {
	logger := bkr.Logger.Session("recreate", lager.Data{
		"instance-id": instanceID,
	})
	callback := bkr.Config.Callbacks.ClusterDataRestore
	if callback == nil {
		err = fmt.Errorf("Broker not configured to support service recreation")
		logger.Error("restore.callback.missing", err, lager.Data{"missing-config": "callbacks.clusterdata_restore"})
		return
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("restore.callback.stdin-pipe", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("restore.callback.stdout-pipe", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("restore.callback.stderr-pipe", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("restore.callback.start", err)
		return
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		defer stdin.Close()
		io.Copy(stdin, bytes.NewBufferString(instanceID))
	}()
	go func() {
		defer wg.Done()
		clusterdata = &serviceinstance.ClusterData{}
		if err := json.NewDecoder(stdout).Decode(&clusterdata); err != nil {
			logger.Error("restore.callback.marshal-error", err)
			return
		}
	}()
	go func() {
		defer wg.Done()
		io.Copy(os.Stderr, stderr)
	}()
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		logger.Error("restore.callback.error", err)
		return
	}
	logger.Info("restore.callback.received", lager.Data{"clusterdata": clusterdata})
	return
}
