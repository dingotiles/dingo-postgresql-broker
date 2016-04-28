package serviceinstance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/dingotiles/dingo-postgresql-broker/bkrconfig"
	"github.com/pivotal-golang/lager"
)

type AdminCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ClusterData describes the current request for the state of the cluster
type ClusterData struct {
	InstanceID       string                 `json:"instance_id"`
	ServiceID        string                 `json:"service_id"`
	PlanID           string                 `json:"plan_id"`
	OrganizationGUID string                 `json:"organization_guid"`
	SpaceGUID        string                 `json:"space_guid"`
	AdminCredentials AdminCredentials       `json:"admin_credentials"`
	Parameters       map[string]interface{} `json:"parameters"`
	NodeCount        int                    `json:"node_count"`
	NodeSize         int                    `json:"node_size"`
	AllocatedPort    string                 `json:"allocated_port"`
}

func (cluster *Cluster) TriggerClusterDataBackup(callbacks bkrconfig.Callbacks) {
	logger := cluster.Logger
	callback := callbacks.ClusterDataBackup
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

func RestoreClusterDataBackup(instanceID string, callbacks bkrconfig.Callbacks, logger lager.Logger) (err error, clusterdata *ClusterData) {
	logger = logger.Session("recreate", lager.Data{
		"instance-id": instanceID,
	})
	callback := callbacks.ClusterDataRestore
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
		clusterdata = &ClusterData{}
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
