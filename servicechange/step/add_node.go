package step

import (
	"fmt"

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
func (step AddNode) Perform(logger lager.Logger) {
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
	// 4. Send requests to backends until one says OK; else fail
	// 5. Store node in KV /clusters/<cluster>/nodes/<node>/backend -> backend uuid
	// 6. Wait until routing mesh allocates public port; and display to logs
	// 7. Return OK; timeout if routing mesh didn't do its job
}
