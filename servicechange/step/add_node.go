package step

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/dingotiles/patroni-broker/config"
	"github.com/dingotiles/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pborman/uuid"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	nodeUUID string
	cluster  *serviceinstance.Cluster
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(cluster *serviceinstance.Cluster) Step {
	return AddNode{cluster: cluster}
}

// StepType prints the type of step
func (step AddNode) StepType() string {
	return "AddNode"
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform() (err error) {
	step.nodeUUID = uuid.New()

	logger := step.cluster.Logger
	logger.Info("add-node.perform", lager.Data{"instance-id": step.cluster.InstanceID, "node-uuid": step.nodeUUID})

	// 1. Generate UUID for node to be created
	// 2. Construct backend provision request (instance_id; service_id, plan_id, org_id, space_id)
	params := step.cluster.ServiceDetails.Parameters
	if params == nil {
		params = map[string]interface{}{}
	}
	params["PATRONI_SCOPE"] = step.cluster.InstanceID
	params["NODE_NAME"] = step.nodeUUID
	provisionDetails := brokerapi.ProvisionDetails{
		OrganizationGUID: step.cluster.ServiceDetails.OrganizationGUID,
		PlanID:           step.cluster.ServiceDetails.PlanID,
		ServiceID:        step.cluster.ServiceDetails.ServiceID,
		SpaceGUID:        step.cluster.ServiceDetails.SpaceGUID,
		Parameters:       params,
	}
	fmt.Println(step.nodeUUID, provisionDetails)

	backends := step.cluster.SortedBackendsByUnusedAZs()
	logger.Info("add-node.perform.backends", lager.Data{
		"backends": backends,
	})

	// 4. Send requests to backends until one says OK; else fail
	var backend *config.Backend
	for _, backend = range backends {
		err = step.requestNodeViaBackend(backend, provisionDetails)
		logger.Error("add-node.perform.try-backend", err)
		if err == nil {
			break
		}
	}
	if err != nil {
		// no backends available to run a cluster
		return err
	}
	// 5. Store node in KV /clusters/<cluster>/nodes/<node>/backend -> backend uuid
	_, err = step.setClusterNodeBackend(backend)
	if err != nil {
		// no backends available to run a cluster
		return err
	}

	// TODO: ensure nodes are in same cluster; I think its currently based on instanceID; but perhaps should be a parameter

	// 6. Return OK; timeout if routing mesh didn't do its job

	return err
}

func (step AddNode) setClusterNodeBackend(backend *config.Backend) (kvIndex uint64, err error) {
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", step.cluster.InstanceID, step.nodeUUID)
	resp, err := step.cluster.EtcdClient.Set(key, backend.GUID, 0)
	if err != nil {
		return 0, err
	}
	return resp.EtcdIndex, err
}

func (step AddNode) requestNodeViaBackend(backend *config.Backend, provisionDetails brokerapi.ProvisionDetails) error {
	var err error
	logger := step.cluster.Logger

	url := fmt.Sprintf("%s/v2/service_instances/%s", backend.URI, step.nodeUUID)
	// client := &http.Client{Timeout: 5}
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	if err = json.NewEncoder(buffer).Encode(provisionDetails); err != nil {
		logger.Error("request-node.backend-provision-encode-details", err)
		return err
	}
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		logger.Error("request-node.backend-provision-req", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(backend.Username, backend.Password)
	if step.cluster.Config.Broker.DumpBackendHTTPTraffic {
		debug(httputil.DumpRequestOut(req, true))
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("request-node.backend-provision-resp", err)
		return err
	}
	if step.cluster.Config.Broker.DumpBackendHTTPTraffic {
		debug(httputil.DumpResponse(resp, true))
	}
	defer resp.Body.Close()

	// FIXME: If resp.StatusCode not 200 or 201, then try next
	if resp.StatusCode >= 400 {
		// FIXME: allow return of this error to end user
		return errors.New("unknown plan")
	}
	return nil
}
