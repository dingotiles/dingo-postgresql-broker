package step

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// RemoveNode instructs cluster to delete a node, starting with replicas
type RemoveNode struct {
	nodeUUID string
	backend  *config.Backend
	cluster  *cluster.Cluster
	logger   lager.Logger
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode(cluster *cluster.Cluster, logger lager.Logger) Step {
	return RemoveNode{cluster: cluster, logger: logger}
}

// StepType prints the type of step
func (step RemoveNode) StepType() string {
	return "RemoveNode"
}

// Perform runs the Step action to modify the Cluster
func (step RemoveNode) Perform() (err error) {
	logger := step.logger

	// 1. Get list of replicas and pick a random one; else pick a random master
	var backendID string
	step.nodeUUID, backendID, err = step.cluster.RandomReplicaNode()
	if err != nil {
		return
	}

	backends := step.cluster.AllBackends()
	for _, backend := range backends {
		if backend.GUID == backendID {
			step.backend = backend
			break
		}
	}
	if step.backend == nil {
		err = fmt.Errorf("Internal error: node assigned to a backend that no longer exists")
		logger.Error("remove-node.perform", err)
		return
	}

	logger.Info("remove-node.perform", lager.Data{
		"instance-id": step.cluster.MetaData().InstanceID,
		"node-uuid":   step.nodeUUID,
		"backend":     step.backend.GUID,
	})

	err = step.requestBackendRemoveNode()
	if err != nil {
		return nil
	}

	err = step.cluster.RemoveNode(step.nodeUUID)
	if err != nil {
		logger.Error("remove-node.nodes-delete", err)
	}
	return
}

func (step RemoveNode) requestBackendRemoveNode() (err error) {
	logger := step.logger

	url := fmt.Sprintf("%s/v2/service_instances/%s", step.backend.URI, step.nodeUUID)
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	deleteDetails := brokerapi.DeprovisionDetails{
		PlanID:    step.cluster.MetaData().PlanID,
		ServiceID: step.cluster.MetaData().ServiceID,
	}

	if err = json.NewEncoder(buffer).Encode(deleteDetails); err != nil {
		logger.Error("remove-node.backend.encode", err)
		return err
	}
	req, err := http.NewRequest("DELETE", url, buffer)
	if err != nil {
		logger.Error("remove-node.backend.new-req", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(step.backend.Username, step.backend.Password)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("remove-node.backend.do", err)
		return err
	}
	defer resp.Body.Close()

	return
}
