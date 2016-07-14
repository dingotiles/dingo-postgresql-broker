package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	clusterModel      *state.ClusterModel
	patroni           *patroni.Patroni
	availableBackends backend.Backends
	logger            lager.Logger
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(clusterModel *state.ClusterModel, patroni *patroni.Patroni,
	availableBackends backend.Backends, logger lager.Logger) Step {
	return AddNode{
		clusterModel:      clusterModel,
		patroni:           patroni,
		availableBackends: availableBackends,
		logger:            logger,
	}
}

// StepType prints the type of step
func (step AddNode) StepType() string {
	return "AddNode"
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) Perform() (err error) {
	logger := step.logger
	logger.Info("add-node.perform", lager.Data{"instance-id": step.clusterModel.InstanceID()})

	nodes := step.clusterModel.Nodes()
	clusterStateData := step.clusterModel.Cluster()

	sortedBackends := prioritizeBackends(nodes, step.availableBackends)
	logger.Info("add-node.perform.sorted-backends", lager.Data{"backends": sortedBackends})

	// 4. Send requests to sortedBackends until one says OK; else fail
	var provisionedNode structs.Node
	for _, backend := range sortedBackends {
		provisionedNode, err = backend.ProvisionNode(clusterStateData, step.logger)
		logBackend := lager.Data{
			"uri":  backend.URI,
			"guid": backend.ID,
			"az":   backend.AvailabilityZone,
		}
		if err == nil {
			logger.Info("add-node.perform.sorted-backends.selected", logBackend)
			break
		} else {
			logger.Error("add-node.perform.sorted-backends.skipped", err, logBackend)
		}
	}
	if err != nil {
		// no sortedBackends available to run a cluster
		logger.Error("add-node.perform.sorted-backends.unavailable", err, lager.Data{"summary": "no backends available to run a container"})
		return err
	}
	// 5. Store node in KV states/<cluster>/nodes/<node>/backend -> backend uuid
	err = step.clusterModel.AddNode(provisionedNode)
	if err != nil {
		logger.Error("add-node.perform.add-node", err, lager.Data{"summary": "no sorted-backends available to run a cluster"})
		return err
	}

	// 6. Wait until node registers itself in data store
	logger.Info("add-node.perform.wait-til-exists", lager.Data{"member": provisionedNode.ID})
	err = step.patroni.WaitForMember(step.clusterModel.InstanceID(), provisionedNode.ID)
	if err != nil {
		logger.Error("add-node.perform.wait-til-exists.error", err, lager.Data{"member": provisionedNode.ID})
		return err
	}

	logger.Info("add-node.perform.success", lager.Data{"member": provisionedNode.ID})
	return nil
}

func prioritizeBackends(existingNodes []*structs.Node, backends backend.Backends) backend.Backends {
	usedBackendIDs := []string{}
	for _, node := range existingNodes {
		usedBackendIDs = append(usedBackendIDs, node.BackendID)
	}
	return sortedBackendsByUnusedAZs(usedBackendIDs, backends)
}

func sortedBackendsByUnusedAZs(usedBackendIDs []string, backends backend.Backends) backend.Backends {
	usedBackends, unusedBackeds := usedAndUnusedBackends(usedBackendIDs, backends)
	ret := backend.Backends{}

	for _, az := range sortBackendAZsByUnusedness(usedBackendIDs, backends).Keys {
		for _, backend := range unusedBackeds {
			if backend.AvailabilityZone == az {
				ret = append(ret, backend)
			}
		}
	}
	for _, backend := range usedBackends {
		ret = append(ret, backend)
	}
	return ret
}

// backendAZsByUnusedness sorts the availability zones in order of whether this cluster is using them or not
// An AZ that is not being used at all will be early in the result.
// All known AZs are included in the result
func sortBackendAZsByUnusedness(usedBackendIDs []string, backends backend.Backends) (vs *utils.ValSorter) {
	azUsageData := map[string]int{}
	for _, az := range backends.AllAvailabilityZones() {
		azUsageData[az] = 0
	}
	for _, backendID := range usedBackendIDs {
		if az, err := backends.AvailabilityZone(backendID); err != nil {
			azUsageData[az]++
		}
	}
	vs = utils.NewValSorter(azUsageData)
	vs.Sort()
	return
}

func usedAndUnusedBackends(usedBackendIDs []string, backends backend.Backends) (usedBackends, unusuedBackends backend.Backends) {
	for _, backend := range backends {
		used := false
		for _, usedBackendID := range usedBackendIDs {
			if backend.ID == usedBackendID {
				usedBackends = append(usedBackends, backend)
				used = true
				break
			}
		}
		if !used {
			unusuedBackends = append(unusuedBackends, backend)
		}
	}
	return
}
