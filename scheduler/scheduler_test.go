package scheduler

import (
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

func TestScheduler_filterBackendsByCellGUIDs(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_filterBackendsByCellGUIDs"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell-guid1"},
			&config.Backend{GUID: "cell-guid2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, logger)
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

	if len(plan.availableBackends) != 1 {
		t.Fatalf("Plan should only have one filtered backend")
	}
	if len(plan.allBackends) != 2 {
		t.Fatalf("Plan should only have two backends")
	}
}

func TestScheduler_allBackends(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_allBackends"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell-guid1"},
			&config.Backend{GUID: "cell-guid2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterModel := state.NewClusterModel(&state.StateEtcd{}, structs.ClusterState{InstanceID: "test"})
	features := structs.ClusterFeatures{}
	plan, err := scheduler.newPlan(clusterModel, features)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}

	if len(plan.availableBackends) != 2 {
		t.Fatalf("Plan should have both backend cells")
	}
	if len(plan.allBackends) != 2 {
		t.Fatalf("Plan should only have two backends")
	}
}

func TestScheduler_VerifyClusterFeatures(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_VerifyClusterFeatures"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler, err := NewScheduler(config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "a"},
			&config.Backend{GUID: "b"},
			&config.Backend{GUID: "c"},
			&config.Backend{GUID: "d"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}, logger)
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
		Backends: []*config.Backend{},
		Etcd:     testutil.LocalEtcdConfig,
	}, logger)
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
