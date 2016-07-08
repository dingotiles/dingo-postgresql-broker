package broker

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

type Callbacks struct {
	backupCallback  *config.CallbackCommand
	restoreCallback *config.CallbackCommand
	logger          lager.Logger
}

func NewCallbacks(config config.Callbacks, logger lager.Logger) *Callbacks {
	callbacks := &Callbacks{
		backupCallback:  config.ClusterDataBackup,
		restoreCallback: config.ClusterDataRestore,
		logger:          logger,
	}
	return callbacks
}

func (c *Callbacks) Configured() bool {
	return c.backupCallback != nil && c.restoreCallback != nil
}

func (c *Callbacks) WriteRecreationData(clusterData *structs.ClusterRecreationData) {
	callback := c.backupCallback
	logger := c.logger

	if callback == nil {
		logger.Info("callbacks.write-data.noop")
		return
	}

	data, err := json.Marshal(clusterData)
	if err != nil {
		logger.Error("callbacks.write-data.data-marshal", err)
		return
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("callbacks.write-data.stdin-pipe", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("callbacks.write-data.stdout-pipe", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("callbacks.write-data.stderr-pipe", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("callbacks.write-data.start", err)
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
		logger.Error("callbacks.write-data.cmd", err)
		return
	}
	logger.Info("callbacks.write-data.done")
}

func (c *Callbacks) RestoreRecreationData(instanceID structs.ClusterID) (*structs.ClusterRecreationData, error) {
	callback := c.restoreCallback
	logger := c.logger

	if callback == nil {
		err := fmt.Errorf("Broker not configured to support service recreation")
		logger.Error("callbacks.restore.callback-missing", err, lager.Data{"missing-config": "callbacks.clusterdata_restore"})
		return nil, err
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("callbacks.restore.stdin-pipe", err)
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("callbacks.restore.stdout-pipe", err)
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("callbacks.restore.stderr-pipe", err)
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("callbacks.restore.start", err)
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		defer stdin.Close()
		io.Copy(stdin, bytes.NewBufferString(string(instanceID)))
	}()
	clusterData := &structs.ClusterRecreationData{}
	go func() {
		defer wg.Done()
		if err := json.NewDecoder(stdout).Decode(&clusterData); err != nil {
			logger.Error("callbacks.restore.marshal-error", err)
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
		logger.Error("callbacks.restore.error", err)
		return nil, err
	}
	logger.Info("callbacks.restore.done", lager.Data{"clusterData": clusterData})
	return clusterData, nil
}
