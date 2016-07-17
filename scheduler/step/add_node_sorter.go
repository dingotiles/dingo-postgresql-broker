package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
)

func (step AddNode) prioritizeCellsToTry(existingNodes []*structs.Node) (sorted backend.Backends, err error) {
	if len(existingNodes) == 0 {
		// Select first node from across all healthy cells irrespective of AZ
		return step.prioritizeCellsByHealth(existingNodes, step.availableBackends)
	}
	usedBackends, unusedBackeds := step.usedAndUnusedBackends(existingNodes, step.availableBackends)

	for _, az := range step.sortBackendAZsByUnusedness(existingNodes, step.availableBackends).Keys {
		unusedBackendsInAZ := backend.Backends{}
		for _, backend := range unusedBackeds {
			if backend.AvailabilityZone == az {
				unusedBackendsInAZ = append(unusedBackendsInAZ, backend)
			}
		}
		sortedUnusedBackendsInAZ, err := step.prioritizeCellsByHealth(existingNodes, unusedBackendsInAZ)
		if err != nil {
			return nil, err
		}
		for _, backend := range sortedUnusedBackendsInAZ {
			sorted = append(sorted, backend)
		}
	}
	for _, backend := range usedBackends {
		sorted = append(sorted, backend)
	}
	return
}

// backendAZsByUnusedness sorts the availability zones in order of whether this cluster is using them or not
// An AZ that is not being used at all will be early in the result.
// All known AZs are included in the result
func (step AddNode) sortBackendAZsByUnusedness(existingNodes []*structs.Node, backends backend.Backends) (vs *utils.ValSorter) {
	azUsageData := map[string]int{}
	for _, az := range backends.AllAvailabilityZones() {
		azUsageData[az] = 0
	}
	for _, existingNode := range existingNodes {
		if az, err := backends.AvailabilityZone(existingNode.BackendID); err == nil {
			azUsageData[az] += 1
		}
	}
	vs = utils.NewValSorter(azUsageData)
	vs.Sort()
	return
}

func (step AddNode) usedAndUnusedBackends(existingNodes []*structs.Node, backends backend.Backends) (usedBackends, unusuedBackends backend.Backends) {
	for _, backend := range backends {
		used := false
		for _, existingNode := range existingNodes {
			if backend.ID == existingNode.BackendID {
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

// Perform runs the Step action to modify the Cluster
func (step AddNode) prioritizeCellsByHealth(existingNodes []*structs.Node, backends backend.Backends) (cellsToTry backend.Backends, err error) {
	// nodes := step.clusterModel.Nodes()
	availableCells := make([]*config.Backend, len(backends))
	for i, cell := range backends {
		availableCells[i] = &config.Backend{GUID: cell.ID}
	}
	// Prioritize availableCells into [unused AZs, used AZs, used cells]
	health, err := backends.InspectHealth()
	if err != nil {
		return
	}
	vs := utils.NewValSorter(health)
	vs.Sort()
	for _, nextCellID := range vs.Keys {
		for _, cellAPI := range backends {
			if cellAPI.ID == nextCellID {
				cellsToTry = append(cellsToTry, cellAPI)
				break
			}
		}
	}

	return
}
