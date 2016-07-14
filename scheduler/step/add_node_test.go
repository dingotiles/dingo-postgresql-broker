package step

import (
	"reflect"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/cells"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

type CellsTest struct {
	health *cells.CellsHealth
}

func (c *CellsTest) LoadStatus(availableCells []*config.Backend) (health *cells.CellsHealth, err error) {
	return c.health, nil
}

func TestAddNode_PrioritizeCells_FirstNode(t *testing.T) {
	t.Parallel()

	testPrefix := "TestAddNode_PrioritizeCells_FirstNode"
	testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	availableCells := backend.Backends{
		&backend.Backend{ID: "cell-n1-z1", AvailabilityZone: "z1"},
		&backend.Backend{ID: "cell-n2-z1", AvailabilityZone: "z1"},
		&backend.Backend{ID: "cell-n3-z2", AvailabilityZone: "z2"},
		&backend.Backend{ID: "cell-n4-z2", AvailabilityZone: "z2"},
	}
	cellsHealth := CellsTest{
		health: &cells.CellsHealth{
			"cell-n1-z1": 3,
			"cell-n2-z1": 2,
			"cell-n3-z2": 1,
			"cell-n4-z2": 0,
		},
	}
	step := AddNode{logger: logger, availableBackends: availableCells, cellsHealth: &cellsHealth}
	cellsToTry, _ := step.prioritizeCellsToTry()
	cellIDs := []string{}
	for _, cell := range cellsToTry {
		cellIDs = append(cellIDs, cell.ID)
	}
	expectedPriority := []string{"cell-n4-z2", "cell-n3-z2", "cell-n2-z1", "cell-n1-z1"}
	if !reflect.DeepEqual(cellIDs, expectedPriority) {
		t.Fatalf("Expected prioritized cells %v to be %v", cellIDs, expectedPriority)
	}

}
