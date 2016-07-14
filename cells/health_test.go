package cells

import (
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

func TestHealth_Load(t *testing.T) {
	t.Parallel()

	testPrefix := "TestHealth_Load"
	testutil.ResetEtcd(t, testPrefix)
	// etcdApi := testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	state, err := state.NewStateEtcdWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create state", err)
	}

	clusterState := structs.ClusterState{
		InstanceID:       "test-1",
		OrganizationGUID: "OrganizationGUID",
		PlanID:           "PlanID",
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
		Nodes: []*structs.Node{
			&structs.Node{BackendID: "backend1-az1"},
			&structs.Node{BackendID: "backend1-az2"},
		},
	}
	err = state.SaveCluster(clusterState)
	if err != nil {
		t.Fatalf("SaveCluster test-1 failed %s", err)
	}

	clusterState = structs.ClusterState{
		InstanceID:       "test-2",
		OrganizationGUID: "OrganizationGUID",
		PlanID:           "PlanID",
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
		Nodes: []*structs.Node{
			&structs.Node{BackendID: "backend1-az1"},
			&structs.Node{BackendID: "backend2-az2"},
		},
	}
	err = state.SaveCluster(clusterState)
	if err != nil {
		t.Fatalf("SaveCluster test-2 failed %s", err)
	}

	cells, err := NewCellsEtcdWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not discover cells from etcd", err)
	}

	cellsStatus, err := cells.LoadStatus()
	if err != nil {
		t.Fatalf("Failed to load cells health", err)
	}
	if count := (*cellsStatus)["backend1-az1"].NodeCount; count != 2 {
		t.Fatalf("Expect cell backend1-az1 to have 2 nodes, found %d", count)
	}
	if count := (*cellsStatus)["backend1-az2"].NodeCount; count != 1 {
		t.Fatalf("Expect cell backend1-az2 to have 1 nodes, found %d", count)
	}
	if count := (*cellsStatus)["backend2-az2"].NodeCount; count != 1 {
		t.Fatalf("Expect cell backend1-az2 to have 1 nodes, found %d", count)
	}
	if (*cellsStatus)["backend2-az1"] != nil {
		t.Fatalf("backend2-az1 has no nodes assigned to it; so should not be included")
	}
}
