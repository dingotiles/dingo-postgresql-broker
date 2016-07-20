package cells

import (
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
)

type FakeClusterLoader struct {
	Clusters []*structs.ClusterState
}

func (f *FakeClusterLoader) LoadAllRunningClusters() ([]*structs.ClusterState, error) {
	return f.Clusters, nil
}

func TestHealth_Load_SomeUnusedCells(t *testing.T) {
	t.Parallel()

	clusterLoader := &FakeClusterLoader{
		Clusters: []*structs.ClusterState{
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell1-az1"},
					&structs.Node{CellGUID: "cell3-az2"},
				}},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell1-az1"},
					&structs.Node{CellGUID: "cell4-az2"},
				}},
		},
	}

	availableCells := []*config.Cell{
		&config.Cell{GUID: "cell1-az1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell2-az1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell3-az2", AvailabilityZone: "z2"},
		&config.Cell{GUID: "cell4-az2", AvailabilityZone: "z2"},
	}

	cells := NewCells(availableCells, clusterLoader)
	cellsStatus, err := cells.InspectHealth()
	if err != nil {
		t.Fatalf("Failed to load cells health", err)
	}

	if health := cellsStatus["cell1-az1"]; health != 2 {
		t.Fatalf("Expect cell cell1-az1 to have health 2, found %d", health)
	}
	if health := cellsStatus["cell3-az2"]; health != 1 {
		t.Fatalf("Expect cell cell3-az2 to have health 1, found %d", health)
	}
	if health := cellsStatus["cell4-az2"]; health != 1 {
		t.Fatalf("Expect cell cell4-az2 to have health 1, found %d", health)
	}
	health, ok := cellsStatus["cell2-az1"]
	if !ok {
		t.Fatalf("cell2-az1 has no nodes assigned to it; but should still be included")
	}
	if health != 0 {
		t.Fatalf("Expect cell cell2-az1 to have health 0, found %d", health)
	}
}

func TestHealth_Load_SubsetAvailableCells(t *testing.T) {
	t.Parallel()

	clusterLoader := &FakeClusterLoader{
		Clusters: []*structs.ClusterState{
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell1-az1"},
					&structs.Node{CellGUID: "cell3-az2"},
				},
			},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{CellGUID: "cell1-az1"},
					&structs.Node{CellGUID: "cell4-az2"},
				},
			},
		},
	}

	// Filter LoadStatus by a subset of available cells (perhaps admin only wants to focus on subet)
	availableCells := []*config.Cell{
		&config.Cell{GUID: "cell2-az1", AvailabilityZone: "z1"},
		&config.Cell{GUID: "cell4-az2", AvailabilityZone: "z2"},
	}

	cells := NewCells(availableCells, clusterLoader)
	cellsStatus, err := cells.InspectHealth()
	if err != nil {
		t.Fatalf("Failed to load cells health", err)
	}

	health, ok := cellsStatus["cell1-az1"]
	if ok {
		t.Fatalf("cell1-az1 should not be an available cell")
	}

	health, ok = cellsStatus["cell2-az1"]
	if !ok {
		t.Fatalf("cell2-az1 has no nodes assigned to it; but should still be included")
	}
	if health != 0 {
		t.Fatalf("Expect cell cell3-az2 to have health 0, found %d", health)
	}

	health, ok = cellsStatus["cell3-az2"]
	if ok {
		t.Fatalf("cell3-az2 should not be an available cell; found %d", health)
	}

	health, ok = cellsStatus["cell4-az2"]
	if !ok {
		t.Fatalf("cell4-az2 has no nodes assigned to it; but should still be included")
	}
	if health != 1 {
		t.Fatalf("Expect cell cell4-az2 to have health 1, found %d", health)
	}
}
