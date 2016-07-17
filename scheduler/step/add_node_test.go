package step

import (
	"reflect"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

type FakeClusterLoader struct {
	Clusters []*structs.ClusterState
}

func (f *FakeClusterLoader) LoadAllClusters() ([]*structs.ClusterState, error) {
	return f.Clusters, nil
}

func TestAddNode_PrioritizeCells_FirstNode(t *testing.T) {
	t.Parallel()

	testPrefix := "TestAddNode_PrioritizeCells_FirstNode"
	testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	clusterLoader := &FakeClusterLoader{
		Clusters: []*structs.ClusterState{
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell-n1-z1"},
					&structs.Node{CellGUID: "cell-n3-z2"},
				},
			},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell-n1-z1"},
					&structs.Node{CellGUID: "cell-n3-z2"},
				},
			},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell-n1-z1"},
					&structs.Node{CellGUID: "cell-n2-z1"},
				},
			},
		},
	}
	availableCells := cells.NewCells([]*config.Cell{
		&config.Cell{GUID: "cell-n1-z1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell-n2-z1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell-n3-z2", AvailabilityZone: "z2"},
		&config.Cell{GUID: "cell-n4-z2", AvailabilityZone: "z2"},
	}, clusterLoader)

	currentClusterNodes := []*structs.Node{}

	step := AddNode{logger: logger, availableCells: availableCells}
	cellsToTry, _ := step.prioritizeCellsToTry(currentClusterNodes)
	cellIDs := []string{}
	for _, cell := range cellsToTry {
		cellIDs = append(cellIDs, cell.GUID)
	}
	expectedPriority := []string{"cell-n4-z2", "cell-n2-z1", "cell-n3-z2", "cell-n1-z1"}
	if !reflect.DeepEqual(cellIDs, expectedPriority) {
		t.Fatalf("Expected prioritized cells %v to be %v", cellIDs, expectedPriority)
	}

}

func TestAddNode_PrioritizeCells_SecondNodeDiffAZ(t *testing.T) {
	t.Parallel()

	testPrefix := "TestAddNode_PrioritizeCells_SecondNodeDiffAZ"
	logger := testutil.NewTestLogger(testPrefix, t)

	clusterLoader := &FakeClusterLoader{
		Clusters: []*structs.ClusterState{
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell-n1-z1"},
					&structs.Node{CellGUID: "cell-n3-z2"},
				},
			},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell-n1-z1"},
					&structs.Node{CellGUID: "cell-n3-z2"},
				},
			},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell-n1-z1"},
					&structs.Node{CellGUID: "cell-n2-z1"},
				},
			},
		},
	}
	availableCells := cells.NewCells([]*config.Cell{
		&config.Cell{GUID: "cell-n1-z1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell-n2-z1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell-n3-z2", AvailabilityZone: "z2"},
		&config.Cell{GUID: "cell-n4-z2", AvailabilityZone: "z2"},
	}, clusterLoader)
	currentClusterNodes := []*structs.Node{
		&structs.Node{ID: "node-1", CellGUID: "cell-n1-z1"},
	}

	step := AddNode{logger: logger, availableCells: availableCells}
	cellsToTry, _ := step.prioritizeCellsToTry(currentClusterNodes)
	cellIDs := []string{}
	for _, cell := range cellsToTry {
		cellIDs = append(cellIDs, cell.GUID)
	}
	// Expect all z2 AZs first, then z1 AZs as node-1 is in z1 already
	expectedPriority := []string{"cell-n4-z2", "cell-n3-z2", "cell-n2-z1", "cell-n1-z1"}
	if !reflect.DeepEqual(cellIDs, expectedPriority) {
		t.Fatalf("Expected prioritized cells %v to be %v", cellIDs, expectedPriority)
	}

}
