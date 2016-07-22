package scheduler

import (
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/fakes"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

func TestScheduler_filterCellsByCellGUIDs(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_filterCellsByCellGUIDs"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell-guid1"},
			&config.Cell{GUID: "cell-guid2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	features := structs.ClusterFeatures{
		CellGUIDs: []string{"cell-guid1", "unknown-cell-guid"},
	}
	clusterModel := state.NewClusterModel(&state.StateEtcd{}, structs.ClusterState{InstanceID: "test"})
	plan, err := scheduler.newPlan(clusterModel, features)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}

	if len(plan.availableCells) != 1 {
		t.Fatalf("Plan should only have one filtered cell")
	}
	if len(plan.allCells) != 2 {
		t.Fatalf("Plan should only have two cells")
	}
}

func TestScheduler_allCells(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_allCells"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell-guid1"},
			&config.Cell{GUID: "cell-guid2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterModel := state.NewClusterModel(&state.StateEtcd{}, structs.ClusterState{InstanceID: "test"})
	features := structs.ClusterFeatures{}
	plan, err := scheduler.newPlan(clusterModel, features)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}

	if len(plan.availableCells) != 2 {
		t.Fatalf("Plan should have both cell cells")
	}
	if len(plan.allCells) != 2 {
		t.Fatalf("Plan should only have two cells")
	}
}

func TestScheduler_VerifyClusterFeatures(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_VerifyClusterFeatures"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler, err := NewScheduler(config.Scheduler{
		Cells: []*config.Cell{
			&config.Cell{GUID: "a"},
			&config.Cell{GUID: "b"},
			&config.Cell{GUID: "c"},
			&config.Cell{GUID: "d"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	features := structs.ClusterFeatures{
		NodeCount: 3,
		CellGUIDs: []string{"a", "b", "c"},
	}
	err = scheduler.VerifyClusterFeatures(features)
	if err != nil {
		t.Fatalf("Cluster features %v should be valid", features)
	}
}

func TestScheduler_VerifyClusterFeatures_UnknownCellGUIDs(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_VerifyClusterFeatures"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler, err := NewScheduler(config.Scheduler{
		Cells: []*config.Cell{},
		Etcd:  testutil.LocalEtcdConfig,
	}, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	features := structs.ClusterFeatures{
		NodeCount: 3,
		CellGUIDs: []string{"a", "b", "c"},
	}
	err = scheduler.VerifyClusterFeatures(features)
	if err == nil {
		t.Fatalf("Expect 'Cell GUIDs do not match available cells' error")
	}
}
