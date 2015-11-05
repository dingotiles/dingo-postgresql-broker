package step

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	"github.com/pborman/uuid"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	nodeUUID string
	cluster  *serviceinstance.Cluster
	logger   lager.Logger
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(cluster *serviceinstance.Cluster, nodeSize int) Step {
	return AddNode{cluster: cluster}
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform() (err error) {
	logger := step.cluster.Logger
	logger.Info("add-step.perform", lager.Data{"implemented": true, "step": fmt.Sprintf("%#v", step)})
	// 1. Generate UUID for node to be created
	step.nodeUUID = uuid.New()
	// 2. Construct backend provision request (instance_id; service_id, plan_id, org_id, space_id)
	provisionDetails := brokerapi.ProvisionDetails{
		OrganizationGUID: step.cluster.ServiceDetails.OrganizationGUID,
		PlanID:           step.cluster.ServiceDetails.PlanID,
		ServiceID:        step.cluster.ServiceDetails.ServiceID,
		SpaceGUID:        step.cluster.ServiceDetails.SpaceGUID,
		Parameters:       step.cluster.ServiceDetails.Parameters,
	}
	fmt.Println(step.nodeUUID, provisionDetails)

	// 3. Randomize backends from available AZs
	// INITIALLY: fixed list from bosh-lite
	backends := []backend.Backend{
		backend.Backend{GUID: "10.244.21.6", URI: "http://54.234.184.115:10006", Username: "containers", Password: "containers"},
		backend.Backend{GUID: "10.244.21.7", URI: "http://54.234.184.115:10007", Username: "containers", Password: "containers"},
		backend.Backend{GUID: "10.244.21.8", URI: "http://54.234.184.115:10008", Username: "containers", Password: "containers"},
	}
	// 4. Send requests to backends until one says OK; else fail
	// INITIALLY: pick one only
	// for _, backend := range backends {
	list := rand.Perm(len(backends))
	var backend backend.Backend
	fmt.Println("random list of backends", list)
	for _, i := range list {
		backend = backends[i]

		err = step.requestNodeViaBackend(backend, provisionDetails)
		if err == nil {
			break
		}
	}
	if err != nil {
		// no backends available to run a cluster
		return err
	}
	// 5. Store node in KV /clusters/<cluster>/nodes/<node>/backend -> backend uuid
	step.setClusterNodeBackend(backend)

	// TODO: ensure nodes are in same cluster; I think its currently based on instanceID; but perhaps should be a parameter

	// 6. Wait until routing mesh allocates public port; and display to logs
	// This requires access to the same etcd used by backend

	// 7. Return OK; timeout if routing mesh didn't do its job

	return nil
}

func (step AddNode) setClusterNodeBackend(backend backend.Backend) {
	key := fmt.Sprintf("/clusters/%s/nodes/%s/backend", step.cluster.InstanceID, step.nodeUUID)
	step.cluster.EtcdClient.Set(key, backend.GUID, 0)
}

func (step AddNode) requestNodeViaBackend(backend backend.Backend, provisionDetails brokerapi.ProvisionDetails) error {
	var err error
	logger := step.logger

	url := fmt.Sprintf("%s/v2/service_instances/%s", backend.URI, step.nodeUUID)
	// client := &http.Client{Timeout: 5}
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	if err = json.NewEncoder(buffer).Encode(provisionDetails); err != nil {
		logger.Error("backend-provision-encode-details", err)
		return err
	}
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		logger.Error("backend-provision-req", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(backend.Username, backend.Password)
	debug(httputil.DumpRequestOut(req, true))

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("backend-provision-resp", err)
		return err
	}
	debug(httputil.DumpResponse(resp, true))
	defer resp.Body.Close()

	// FIXME: If resp.StatusCode not 200 or 201, then try next
	if resp.StatusCode >= 400 {
		// FIXME: allow return of this error to end user
		return errors.New("unknown plan")
	}
	return nil
}
