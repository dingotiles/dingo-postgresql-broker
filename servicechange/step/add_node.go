package step

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/frodenas/brokerapi"
	"github.com/pborman/uuid"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	serviceDetails brokerapi.ProvisionDetails
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(serviceDetails brokerapi.ProvisionDetails, nodeSize int) Step {
	return AddNode{serviceDetails: serviceDetails}
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform(logger lager.Logger) error {
	logger.Info("add-step.perform", lager.Data{"implemented": true, "step": fmt.Sprintf("%#v", step)})
	// 1. Generate UUID for node to be created
	nodeUUID := uuid.New()
	// 2. Construct backend provision request (instance_id; service_id, plan_id, org_id, space_id)
	provisionDetails := brokerapi.ProvisionDetails{
		OrganizationGUID: step.serviceDetails.OrganizationGUID,
		PlanID:           step.serviceDetails.PlanID,
		ServiceID:        step.serviceDetails.ServiceID,
		SpaceGUID:        step.serviceDetails.SpaceGUID,
		Parameters:       step.serviceDetails.Parameters,
	}
	fmt.Println(nodeUUID, provisionDetails)

	// 3. Randomize backends from available AZs
	// INITIALLY: fixed list from bosh-lite
	backends := []backend.Backend{
		backend.Backend{GUID: "10.244.21.6", URI: "http://10.244.21.6", Username: "containers", Password: "containers"},
		backend.Backend{GUID: "10.244.21.7", URI: "http://10.244.21.7", Username: "containers", Password: "containers"},
		backend.Backend{GUID: "10.244.21.8", URI: "http://10.244.21.8", Username: "containers", Password: "containers"},
	}
	// 4. Send requests to backends until one says OK; else fail
	// INITIALLY: pick one only
	// for _, backend := range backends {
	backend := backends[0]
	var err error

	url := fmt.Sprintf("%s/v2/service_instances/%s", backend.URI, nodeUUID)
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

	// 5. Store node in KV /clusters/<cluster>/nodes/<node>/backend -> backend uuid
	// 6. Wait until routing mesh allocates public port; and display to logs
	// 7. Return OK; timeout if routing mesh didn't do its job
	return nil
}
