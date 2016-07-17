package step

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
)

func (step AddNode) prioritizeCellsToTry(existingNodes []*structs.Node) (sorted cells.Cells, err error) {
	if len(existingNodes) == 0 {
		// Select first node from across all healthy cells irrespective of AZ
		return step.prioritizeCellsByHealth(existingNodes, step.availableCells)
	}
	usedCells, unusedCells := step.usedAndUnusedCells(existingNodes, step.availableCells)

	for _, az := range step.sortCellAZsByUnusedness(existingNodes, step.availableCells).Keys {
		unusedCellsInAZ := cells.Cells{}
		for _, cell := range unusedCells {
			if cell.AvailabilityZone == az {
				unusedCellsInAZ = append(unusedCellsInAZ, cell)
			}
		}
		sortedUnusedCellsInAZ, err := step.prioritizeCellsByHealth(existingNodes, unusedCellsInAZ)
		if err != nil {
			return nil, err
		}
		for _, cell := range sortedUnusedCellsInAZ {
			sorted = append(sorted, cell)
		}
	}
	for _, cell := range usedCells {
		sorted = append(sorted, cell)
	}
	return
}

// CellAZsByUnusednes sorts the availability zones in order of whether this cluster is using them or not
// An AZ that is not being used at all will be early in the result.
// All known AZs are included in the result
func (step AddNode) sortCellAZsByUnusedness(existingNodes []*structs.Node, cells cells.Cells) (vs *utils.ValSorter) {
	azUsageData := map[string]int{}
	for _, az := range cells.AllAvailabilityZones() {
		azUsageData[az] = 0
	}
	for _, existingNode := range existingNodes {
		if az, err := cells.AvailabilityZone(existingNode.CellGUID); err == nil {
			azUsageData[az] += 1
		}
	}
	vs = utils.NewValSorter(azUsageData)
	vs.Sort()
	return
}

func (step AddNode) usedAndUnusedCells(existingNodes []*structs.Node, cells cells.Cells) (usedCells, unusuedCells cells.Cells) {
	for _, cell := range cells {
		used := false
		for _, existingNode := range existingNodes {
			if cell.GUID == existingNode.CellGUID {
				usedCells = append(usedCells, cell)
				used = true
				break
			}
		}
		if !used {
			unusuedCells = append(unusuedCells, cell)
		}
	}
	return
}

// Perform runs the Step action to modify the Cluster
func (step AddNode) prioritizeCellsByHealth(existingNodes []*structs.Node, cells cells.Cells) (cellsToTry cells.Cells, err error) {
	// Prioritize availableCells into [unused AZs, used AZs, used cells]
	health, err := cells.InspectHealth()
	if err != nil {
		return
	}
	vs := utils.NewValSorter(health)
	vs.Sort()
	for _, nextCellID := range vs.Keys {
		for _, cellAPI := range cells {
			if cellAPI.GUID == nextCellID {
				cellsToTry = append(cellsToTry, cellAPI)
				break
			}
		}
	}

	return
}
