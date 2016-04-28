package serviceinstance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
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

func (data *ClusterData) Equals(other *ClusterData) bool {
	return reflect.DeepEqual(*data, *other)
}
func (cluster *Cluster) TriggerClusterDataBackup(callbacks bkrconfig.Callbacks) {
	logger := cluster.Logger
	callback := callbacks.ClusterDataBackup
	if callback == nil {
		logger.Info("clusterdata.backup.noop")
		return
	}

	data, err := json.Marshal(cluster.ClusterData())
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

func RestoreClusterDataBackup(instanceID string, callbacks bkrconfig.Callbacks, logger lager.Logger) (err error, clusterdata *ClusterData) {
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
		clusterdata = &ClusterData{}
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
