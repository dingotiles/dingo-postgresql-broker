package backend

import (
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
)

type FakeClusterLoader struct {
	Clusters []*structs.ClusterState
}

func (f *FakeClusterLoader) LoadAllClusters() ([]*structs.ClusterState, error) {
	return f.Clusters, nil
}

func TestHealth_Load_SomeUnusedCells(t *testing.T) {
	t.Parallel()

	clusterLoader := &FakeClusterLoader{
		Clusters: []*structs.ClusterState{
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{BackendID: "backend1-az1"},
					&structs.Node{BackendID: "backend3-az2"},
				}},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{BackendID: "backend1-az1"},
					&structs.Node{BackendID: "backend4-az2"},
				}},
		},
	}

	availableCells := []*config.Backend{
		&config.Backend{GUID: "backend1-az1", AvailabilityZone: "z1"},
		&config.Backend{GUID: "backend2-az1", AvailabilityZone: "z1"},
		&config.Backend{GUID: "backend3-az2", AvailabilityZone: "z2"},
		&config.Backend{GUID: "backend4-az2", AvailabilityZone: "z2"},
	}

	backends := NewBackends(availableCells, clusterLoader)
	cellsStatus, err := backends.InspectHealth()
	if err != nil {
		t.Fatalf("Failed to load cells health", err)
	}

	if health := cellsStatus["backend1-az1"]; health != 2 {
		t.Fatalf("Expect cell backend1-az1 to have health 2, found %d", health)
	}
	if health := cellsStatus["backend3-az2"]; health != 1 {
		t.Fatalf("Expect cell backend3-az2 to have health 1, found %d", health)
	}
	if health := cellsStatus["backend4-az2"]; health != 1 {
		t.Fatalf("Expect cell backend4-az2 to have health 1, found %d", health)
	}
	health, ok := cellsStatus["backend2-az1"]
	if !ok {
		t.Fatalf("backend2-az1 has no nodes assigned to it; but should still be included")
	}
	if health != 0 {
		t.Fatalf("Expect cell backend2-az1 to have health 0, found %d", health)
	}
}

func TestHealth_Load_SubsetAvailableCells(t *testing.T) {
	t.Parallel()

	clusterLoader := &FakeClusterLoader{
		Clusters: []*structs.ClusterState{
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{BackendID: "backend1-az1"},
					&structs.Node{BackendID: "backend3-az2"},
				},
			},
			&structs.ClusterState{
				Nodes: []*structs.Node{
					&structs.Node{BackendID: "backend1-az1"},
					&structs.Node{BackendID: "backend4-az2"},
				},
			},
		},
	}

	// Filter LoadStatus by a subset of available cells (perhaps admin only wants to focus on subet)
	availableCells := []*config.Backend{
		&config.Backend{GUID: "backend2-az1", AvailabilityZone: "z1"},
		&config.Backend{GUID: "backend4-az2", AvailabilityZone: "z2"},
	}

	backends := NewBackends(availableCells, clusterLoader)
	cellsStatus, err := backends.InspectHealth()
	if err != nil {
		t.Fatalf("Failed to load cells health", err)
	}

	health, ok := cellsStatus["backend1-az1"]
	if ok {
		t.Fatalf("backend1-az1 should not be an available cell")
	}

	health, ok = cellsStatus["backend2-az1"]
	if !ok {
		t.Fatalf("backend2-az1 has no nodes assigned to it; but should still be included")
	}
	if health != 0 {
		t.Fatalf("Expect cell backend3-az2 to have health 0, found %d", health)
	}

	health, ok = cellsStatus["backend3-az2"]
	if ok {
		t.Fatalf("backend3-az2 should not be an available cell; found %d", health)
	}

	health, ok = cellsStatus["backend4-az2"]
	if !ok {
		t.Fatalf("backend4-az2 has no nodes assigned to it; but should still be included")
	}
	if health != 1 {
		t.Fatalf("Expect cell backend4-az2 to have health 1, found %d", health)
	}
}
