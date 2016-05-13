package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

func TriggerClusterDataBackup(clusterData structs.ClusterData, callbacks config.Callbacks, logger lager.Logger) {
	callback := callbacks.ClusterDataBackup
	if callback == nil {
		logger.Info("clusterdata.backup.noop")
		return
	}

	data, err := json.Marshal(clusterData)
	if err != nil {
		logger.Error("clusterdata.backup.data-marshal", err)
		return
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("clusterdata.backup.stdin-pipe", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("clusterdata.backup.stdout-pipe", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("clusterdata.backup.stderr-pipe", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("clusterdata.backup.start", err)
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
		logger.Error("clusterdata.backup.cmd", err)
		return
	}
	logger.Info("clusterdata.backup.done")
}

func RestoreClusterDataBackup(instanceID string, callbacks config.Callbacks, logger lager.Logger) (err error, clusterdata *structs.ClusterData) {
	callback := callbacks.ClusterDataRestore
	if callback == nil {
		err = fmt.Errorf("Broker not configured to support service recreation")
		logger.Error("clusterdata.restore.callback-missing", err, lager.Data{"missing-config": "callbacks.clusterdata_restore"})
		return
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("clusterdata.restore.stdin-pipe", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("clusterdata.restore.stdout-pipe", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("clusterdata.restore.stderr-pipe", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("clusterdata.restore.start", err)
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
		clusterdata = &structs.ClusterData{}
		if err := json.NewDecoder(stdout).Decode(&clusterdata); err != nil {
			logger.Error("clusterdata.restore.marshal-error", err)
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
		logger.Error("clusterdata.restore.error", err)
		return
	}
	logger.Info("clusterdata.restore.done", lager.Data{"clusterdata": clusterdata})
	return
}
