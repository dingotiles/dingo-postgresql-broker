package scheduler

import (
	"reflect"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

func TestPlan_Steps_NewCluster_Default(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster"
	logger := testutil.NewTestLogger(testPrefix, t)

	config := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
			&config.Backend{GUID: "cell3"},
			&config.Backend{GUID: "cell4"},
		},
	}
	scheduler := NewScheduler(config, logger)
	plan, err := scheduler.newPlan(&structs.ClusterState{}, structs.ClusterFeatures{NodeCount: 2})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "AddNode"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}
