package step

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
	"github.com/pivotal-golang/lager"
)

// AddNode instructs a new cluster node be added
type AddNode struct {
	cluster  *state.Cluster
	backends backend.Backends
	logger   lager.Logger
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
	logger := step.logger
	logger.Info("add-node.perform", lager.Data{"instance-id": step.cluster.MetaData().InstanceID})

	nodes := step.cluster.Nodes()
	sortedBackends := prioritizeBackends(nodes, step.backends)
	logger.Info("add-node.perform.sortedBackends", lager.Data{
		"sortedBackends": sortedBackends,
	})

	// 4. Send requests to sortedBackends until one says OK; else fail
	var provisionedNode state.Node
	for _, backend := range sortedBackends {
		provisionedNode, err = backend.ProvisionNode(step.cluster.MetaData(), step.logger)
		// nodeId, err = step.requestNodeViaBackend(backend, provisionDetails)
		logBackend := lager.Data{
			"uri":  backend.URI,
			"guid": backend.Id,
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
	err = step.cluster.AddNode(provisionedNode)
	if err != nil {
		// no sortedBackends available to run a cluster
		return err
	}

	// TODO: ensure nodes are in same cluster; I think its currently based on instanceID; but perhaps should be a parameter

	// 6. Return OK; timeout if routing mesh didn't do its job

	return err
}

func prioritizeBackends(existingNodes []*state.Node, backends backend.Backends) backend.Backends {
	usedBackendIds := []string{}
	for _, node := range existingNodes {
		usedBackendIds = append(usedBackendIds, node.BackendId)
	}
	return sortedBackendsByUnusedAZs(usedBackendIds, backends)
}

func sortedBackendsByUnusedAZs(usedBackendIds []string, backends backend.Backends) backend.Backends {
	usedBackends, unusedBackeds := usedAndUnusedBackends(usedBackendIds, backends)
	ret := backend.Backends{}

	for _, az := range sortBackendAZsByUnusedness(usedBackendIds, backends).Keys {
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
func sortBackendAZsByUnusedness(usedBackendIds []string, backends backend.Backends) (vs *utils.ValSorter) {
	azUsageData := map[string]int{}
	for _, az := range backends.AllAvailabilityZones() {
		azUsageData[az] = 0
	}
	for _, backendId := range usedBackendIds {
		if az, err := backends.AvailabilityZone(backendId); err != nil {
			azUsageData[az]++
		}
	}
	vs = utils.NewValSorter(azUsageData)
	fmt.Printf("usage %#v\n", azUsageData)
	vs.Sort()
	fmt.Printf("sorted %#v\n", vs)
	return
}

func usedAndUnusedBackends(usedBackendIds []string, backends backend.Backends) (usedBackends, unusuedBackends backend.Backends) {
	for _, backend := range backends {
		used := false
		for _, usedBackendId := range usedBackendIds {
			if backend.Id == usedBackendId {
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
