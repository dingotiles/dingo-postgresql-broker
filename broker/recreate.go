package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Recreate(instanceID string, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	logger := bkr.Logger.Session("recreate", lager.Data{
		"instance-id": instanceID,
	})

	logger.Info("start", lager.Data{})
	err, _ = bkr.restoreClusterDataBackup(instanceID)

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
