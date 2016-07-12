package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/patronidata"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	cluster           *structs.ClusterState
	clusterData       patronidata.ClusterDataWrapper
	availableBackends backend.Backends
	logger            lager.Logger
}

// NewStepAddNode creates a StepAddNode command
func NewStepAddNode(cluster *structs.ClusterState, clusterData patronidata.ClusterDataWrapper,
	availableBackends backend.Backends, logger lager.Logger) Step {
	return AddNode{
		cluster:           cluster,
		clusterData:       clusterData,
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
	logger.Info("add-node.perform", lager.Data{"instance-id": step.cluster.InstanceID})

	nodes := step.cluster.Nodes
	sortedBackends := prioritizeBackends(nodes, step.availableBackends)
	logger.Info("add-node.perform.sorted-backends", lager.Data{
		"sortedBackends": sortedBackends,
	})

	// 4. Send requests to sortedBackends until one says OK; else fail
	var provisionedNode structs.Node
	for _, backend := range sortedBackends {
		provisionedNode, err = backend.ProvisionNode(step.cluster, step.logger)
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
	err = step.cluster.AddNode(provisionedNode)
	if err != nil {
		logger.Error("add-node.perform.add-node.error", err, lager.Data{"summary": "no sorted-backends available to run a cluster"})
		return err
	}

	// 6. Block until node is state == running
	err = step.clusterData.WaitTilMemberRunning(provisionedNode.ID)
	if err != nil {
		logger.Error("add-node.perform.wait-til-running.error", err, lager.Data{"member": provisionedNode.ID})
		return err
	}

	return err
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
