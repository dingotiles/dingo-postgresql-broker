package step

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// RemoveNode instructs cluster to delete a node, starting with replicas
type RemoveNode struct {
	nodeUUID string
	backend  *backend.Backend
	cluster  *serviceinstance.Cluster
}

// NewStepRemoveNode creates a StepRemoveNode command
func NewStepRemoveNode(cluster *serviceinstance.Cluster) Step {
	return RemoveNode{cluster: cluster}
}

// Perform runs the Step action to modify the Cluster
func (step RemoveNode) Perform() (err error) {
	logger := step.cluster.Logger

	// 1. Get list of replicas and pick a random one; else pick a random master
	var backendID string
	step.nodeUUID, backendID, err = step.cluster.RandomReplicaNode()
	if err != nil {
		return
	}

	backends := []backend.Backend{
		backend.Backend{GUID: "10.244.21.6", URI: "http://54.145.50.109:10006", Username: "containers", Password: "containers"},
		backend.Backend{GUID: "10.244.21.7", URI: "http://54.145.50.109:10007", Username: "containers", Password: "containers"},
		backend.Backend{GUID: "10.244.21.8", URI: "http://54.145.50.109:10008", Username: "containers", Password: "containers"},
	}
	for _, backend := range backends {
		if backend.GUID == backendID {
			step.backend = &backend
			break
		}
	}
	if step.backend == nil {
		err = fmt.Errorf("Internal error: node assigned to a backend that no longer exists")
		logger.Error("remove-node.perform", err)
		return
	}

	logger.Info("remove-node.perform", lager.Data{
		"instance-id": step.cluster.InstanceID,
		"node-uuid":   step.nodeUUID,
		"backend":     step.backend.GUID,
	})

	err = step.requestBackendRemoveNode()
	if err != nil {
		return nil
	}

	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s", step.cluster.InstanceID, step.nodeUUID)
	_, err = step.cluster.EtcdClient.Delete(key, true)
	if err != nil {
		logger.Error("remove-node.nodes-delete", err)
	}
	return
}

func (step RemoveNode) requestBackendRemoveNode() (err error) {
	logger := step.cluster.Logger

	url := fmt.Sprintf("%s/v2/service_instances/%s", step.backend.URI, step.nodeUUID)
	// client := &http.Client{Timeout: 5}
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	deleteDetails := brokerapi.DeprovisionDetails{
		PlanID:    step.cluster.ServiceDetails.PlanID,
		ServiceID: step.cluster.ServiceDetails.ServiceID,
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
	debug(httputil.DumpRequestOut(req, true))

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("remove-node.backend.do", err)
		return err
	}
	debug(httputil.DumpResponse(resp, true))
	defer resp.Body.Close()

	return
}
