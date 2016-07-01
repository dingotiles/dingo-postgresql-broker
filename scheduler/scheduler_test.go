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
		CellGUIDsForNewNodes: []string{"cell-guid1", "unknown-cell-guid"},
	}
	plan, err := scheduler.newPlan(nil, features)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}

	if len(plan.backends) != 1 {
		t.Fatalf("Plan should only have one filtered backend")
	}
}

func TestScheduler_allBackends(t *testing.T) {
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
	features := structs.ClusterFeatures{}
	plan, err := scheduler.newPlan(nil, features)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}

	if len(plan.backends) != 2 {
		t.Fatalf("Plan should have both backend cells")
	}
}
