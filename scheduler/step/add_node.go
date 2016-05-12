package step

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/frodenas/brokerapi"
	"github.com/pborman/uuid"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	cluster  *state.Cluster
	backends backend.Backends
	logger   lager.Logger
	nodeUUID string
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(cluster *state.Cluster, backends backend.Backends, logger lager.Logger) Step {
	return AddNode{cluster: cluster, backends: backends, logger: logger}
}

// StepType prints the type of step
func (step AddNode) StepType() string {
	return "AddNode"
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform() (err error) {
	var backends []*config.Backend
	for _, b := range step.backends {
		backends = append(backends, b.Config)
	}

	step.nodeUUID = uuid.New()

	logger := step.logger
	logger.Info("add-node.perform", lager.Data{"instance-id": step.cluster.MetaData().InstanceID, "node-uuid": step.nodeUUID})

	// 1. Generate UUID for node to be created
	// 2. Construct backend provision request (instance_id; service_id, plan_id, org_id, space_id)
	params := map[string]interface{}{}
	params["PATRONI_SCOPE"] = step.cluster.MetaData().InstanceID
	params["NODE_NAME"] = step.nodeUUID
	params["POSTGRES_USERNAME"] = step.cluster.MetaData().AdminCredentials.Username
	params["POSTGRES_PASSWORD"] = step.cluster.MetaData().AdminCredentials.Password
	provisionDetails := brokerapi.ProvisionDetails{
		OrganizationGUID: step.cluster.MetaData().OrganizationGUID,
		PlanID:           step.cluster.MetaData().PlanID,
		ServiceID:        step.cluster.MetaData().ServiceID,
		SpaceGUID:        step.cluster.MetaData().SpaceGUID,
		Parameters:       params,
	}
	fmt.Println(step.nodeUUID, provisionDetails)

	sortedBackends := step.cluster.SortedBackendsByUnusedAZs(backends)
	logger.Info("add-node.perform.sortedBackends", lager.Data{
		"sortedBackends": sortedBackends,
	})

	// 4. Send requests to sortedBackends until one says OK; else fail
	var backend *config.Backend
	for _, backend = range sortedBackends {
		err = step.requestNodeViaBackend(backend, provisionDetails)
		logBackend := lager.Data{
			"uri":  backend.URI,
			"guid": backend.GUID,
			"az":   backend.AvailabilityZone,
		}
		if err == nil {
			logger.Info("add-node.perform.sortedBackends.selected", logBackend)
			break
		} else {
			logger.Error("add-node.perform.sortedBackends.skipped", err, logBackend)
		}
	}
	if err != nil {
		// no sortedBackends available to run a cluster
		logger.Info("add-node.perform.sortedBackends.unavailable", lager.Data{"summary": "no backends available to run a container"})
		return err
	}
	// 5. Store node in KV states/<cluster>/nodes/<node>/backend -> backend uuid
	_, err = step.setClusterNodeBackend(backend)
	if err != nil {
		// no sortedBackends available to run a cluster
		return err
	}

	// TODO: ensure nodes are in same cluster; I think its currently based on instanceID; but perhaps should be a parameter

	// 6. Return OK; timeout if routing mesh didn't do its job

	return err
}

func (step AddNode) setClusterNodeBackend(backend *config.Backend) (kvIndex uint64, err error) {
	resp, err := step.cluster.AddNode(state.Node{Id: step.nodeUUID, BackendId: backend.GUID})
	if err != nil {
		return 0, err
	}
	return resp, err
}

func (step AddNode) requestNodeViaBackend(backend *config.Backend, provisionDetails brokerapi.ProvisionDetails) error {
	var err error
	logger := step.logger

	url := fmt.Sprintf("%s/v2/service_instances/%s", backend.URI, step.nodeUUID)
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

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("request-node.backend-provision-resp", err)
		return err
	}
	defer resp.Body.Close()

	// FIXME: If resp.StatusCode not 200 or 201, then try next
	if resp.StatusCode >= 400 {
		// FIXME: allow return of this error to end user
		return errors.New("unknown plan")
	}
	return nil
}
