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
	backupCallback     *config.CallbackCommand
	restoreCallback    *config.CallbackCommand
	findByNameCallback *config.CallbackCommand
	logger             lager.Logger
}

func NewCallbacks(config config.Callbacks, logger lager.Logger) *Callbacks {
	callbacks := &Callbacks{
		backupCallback:     config.ClusterDataBackup,
		restoreCallback:    config.ClusterDataRestore,
		findByNameCallback: config.ClusterDataFindByName,
		logger:             logger,
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

// Input: {"space_guid": "GUID", "name": "NAME"}
// Output: {"instance_id":"71c27bbe-...", ...}
func (c *Callbacks) ClusterDataFindServiceInstanceByName(spaceGUID, name string) (*structs.ClusterRecreationData, error) {
	callback := c.findByNameCallback
	logger := c.logger

	if callback == nil {
		err := fmt.Errorf("Broker not configured to support discovery of existing clusterdata backups by name")
		logger.Error("callbacks.find-by-name.callback-missing", err, lager.Data{"missing-config": "callbacks.clusterdata_find_by_name"})
		return nil, err
	}

	data, err := json.Marshal(map[string]string{
		"space_guid": spaceGUID,
		"name":       name,
	})
	if err != nil {
		logger.Error("callbacks.write-data.data-marshal", err)
		return nil, err
	}

	cmd := exec.Command(callback.Command, callback.Arguments...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("callbacks.find-by-name.stdin-pipe", err)
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("callbacks.find-by-name.stdout-pipe", err)
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("callbacks.find-by-name.stderr-pipe", err)
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		logger.Error("callbacks.find-by-name.start", err)
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		defer stdin.Close()
		io.Copy(stdin, bytes.NewBufferString(string(data)))
	}()
	clusterData := &structs.ClusterRecreationData{}
	go func() {
		defer wg.Done()
		if err = json.NewDecoder(stdout).Decode(&clusterData); err != nil {
			logger.Error("callbacks.find-by-name.marshal-error", err)
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
		logger.Error("callbacks.find-by-name.error", err)
		return nil, err
	}
	if clusterData.InstanceID == "" {
		return nil, fmt.Errorf("Failed to fetch backup clusterdata for %s / %s", spaceGUID, name)
	}
	logger.Info("callbacks.find-by-name.done", lager.Data{"clusterData": clusterData})
	return clusterData, nil
}
