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

	testPrefix := "TestPlan_Steps_NewCluster_Default"
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

func TestPlan_Steps_NewCluster_IncreaseCount(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster_IncreaseCount"
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
	clusterState := &structs.ClusterState{
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell1"},
			&structs.Node{ID: "b", BackendID: "cell2"},
		},
	}
	plan, err := scheduler.newPlan(clusterState, structs.ClusterFeatures{NodeCount: 3})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}

func TestPlan_Steps_NewCluster_DecreaseCount(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster_DecreaseCount"
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
	clusterState := &structs.ClusterState{
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell1"},
			&structs.Node{ID: "b", BackendID: "cell2"},
			&structs.Node{ID: "c", BackendID: "cell3"},
		},
	}
	plan, err := scheduler.newPlan(clusterState, structs.ClusterFeatures{NodeCount: 2})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"RemoveNode"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}
