package scheduler

import (
	"reflect"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
)

func TestPlan_Steps_NewCluster_Default(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster_Default"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
			&config.Backend{GUID: "cell3"},
			&config.Backend{GUID: "cell4"},
		},
	}
	scheduler := NewScheduler(schedulerConfig, logger)
	plan, err := scheduler.newPlan(&structs.ClusterState{}, testutil.LocalEtcdConfig, structs.ClusterFeatures{NodeCount: 2})
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

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
			&config.Backend{GUID: "cell3"},
			&config.Backend{GUID: "cell4"},
		},
	}
	scheduler := NewScheduler(schedulerConfig, logger)
	clusterState := &structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell1", Role: state.LeaderRole},
			&structs.Node{ID: "b", BackendID: "cell2", Role: state.ReplicaRole},
		},
	}
	plan, err := scheduler.newPlan(clusterState, testutil.LocalEtcdConfig, structs.ClusterFeatures{NodeCount: 3})
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

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
			&config.Backend{GUID: "cell3"},
			&config.Backend{GUID: "cell4"},
		},
	}
	scheduler := NewScheduler(schedulerConfig, logger)
	clusterState := &structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell1", Role: state.LeaderRole},
			&structs.Node{ID: "b", BackendID: "cell2", Role: state.ReplicaRole},
			&structs.Node{ID: "c", BackendID: "cell3", Role: state.ReplicaRole},
		},
	}
	plan, err := scheduler.newPlan(clusterState, testutil.LocalEtcdConfig, structs.ClusterFeatures{NodeCount: 2})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"RemoveRandomNode"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}

func TestPlan_Steps_NewCluster_MoveReplica(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster_MoveReplica"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
		},
	}
	scheduler := NewScheduler(schedulerConfig, logger)
	clusterState := &structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell1", Role: state.LeaderRole},
			&structs.Node{ID: "b", BackendID: "cell-unavailable", Role: state.ReplicaRole},
		},
	}
	clusterFeatures := structs.ClusterFeatures{
		NodeCount: 2,
		CellGUIDs: []string{"cell1", "cell2"},
	}
	plan, err := scheduler.newPlan(clusterState, testutil.LocalEtcdConfig, clusterFeatures)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "RemoveNode(b)"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}

func TestPlan_Steps_NewCluster_MoveLeader(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster_MoveLeader"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
		},
	}
	scheduler := NewScheduler(schedulerConfig, logger)
	clusterState := &structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell-unavailable", Role: state.LeaderRole},
			&structs.Node{ID: "b", BackendID: "cell2", Role: state.ReplicaRole},
		},
	}
	clusterFeatures := structs.ClusterFeatures{
		NodeCount: 2,
		CellGUIDs: []string{"cell1", "cell2"},
	}
	plan, err := scheduler.newPlan(clusterState, testutil.LocalEtcdConfig, clusterFeatures)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "RemoveLeader(a)"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}

func TestPlan_Steps_NewCluster_MoveEverything(t *testing.T) {
	t.Parallel()

	testPrefix := "TestPlan_Steps_NewCluster_MoveEverything"
	logger := testutil.NewTestLogger(testPrefix, t)

	schedulerConfig := config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "cell1"},
			&config.Backend{GUID: "cell2"},
		},
	}
	scheduler := NewScheduler(schedulerConfig, logger)
	clusterState := &structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", BackendID: "cell-x-unavailable", Role: state.LeaderRole},
			&structs.Node{ID: "b", BackendID: "cell-y-unavailable", Role: state.ReplicaRole},
		},
	}
	clusterFeatures := structs.ClusterFeatures{
		NodeCount: 2,
		CellGUIDs: []string{"cell1", "cell2"},
	}
	plan, err := scheduler.newPlan(clusterState, testutil.LocalEtcdConfig, clusterFeatures)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "AddNode", "RemoveNode(b)", "RemoveLeader(a)"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}
