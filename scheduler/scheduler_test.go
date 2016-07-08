package scheduler

import (
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

func TestScheduler_filterBackendsByCellGUIDs(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_filterBackendsByCellGUIDs"
	logger := testutil.NewTestLogger(testPrefix, t)

	config := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell-guid1"},
			&config.Backend{GUID: "cell-guid2"},
		},
	}
	scheduler := NewScheduler(config, logger)
	features := structs.ClusterFeatures{
		CellGUIDs: []string{"cell-guid1", "unknown-cell-guid"},
	}
	plan, err := scheduler.newPlan(nil, features)
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

	config := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell-guid1"},
			&config.Backend{GUID: "cell-guid2"},
		},
	}
	scheduler := NewScheduler(config, logger)
	features := structs.ClusterFeatures{}
	plan, err := scheduler.newPlan(nil, features)
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
	scheduler := NewScheduler(config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "a"},
			&config.Backend{GUID: "b"},
			&config.Backend{GUID: "c"},
			&config.Backend{GUID: "d"},
		},
	}, logger)
	features := structs.ClusterFeatures{
		NodeCount: 3,
		CellGUIDs: []string{"a", "b", "c"},
	}
	err := scheduler.VerifyClusterFeatures(features)
	if err != nil {
		t.Fatalf("Cluster features %v should be valid", features)
	}
}

func TestScheduler_VerifyClusterFeatures_UnknownCellGUIDs(t *testing.T) {
	t.Parallel()

	testPrefix := "TestScheduler_VerifyClusterFeatures"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler := NewScheduler(config.Scheduler{
		Backends: []*config.Backend{},
	}, logger)
	features := structs.ClusterFeatures{
		NodeCount: 3,
		CellGUIDs: []string{"a", "b", "c"},
	}
	err := scheduler.VerifyClusterFeatures(features)
	if err == nil {
		t.Fatalf("Expect 'Cell GUIDs do not match available cells' error")
	}
}
