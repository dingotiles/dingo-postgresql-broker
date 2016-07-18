package scheduler

import (
	"reflect"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/fakes"
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
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell1"},
			&config.Cell{GUID: "cell2"},
			&config.Cell{GUID: "cell3"},
			&config.Cell{GUID: "cell4"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterModel := state.NewClusterModel(&state.StateEtcd{}, structs.ClusterState{})
	plan, err := scheduler.newPlan(clusterModel, structs.ClusterFeatures{NodeCount: 2})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "AddNode", "WaitForAllMembers", "WaitForLeader"}
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
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell1"},
			&config.Cell{GUID: "cell2"},
			&config.Cell{GUID: "cell3"},
			&config.Cell{GUID: "cell4"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterState := structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", CellGUID: "cell1", Role: state.LeaderRole},
			&structs.Node{ID: "b", CellGUID: "cell2", Role: state.ReplicaRole},
		},
	}
	clusterModel := state.NewClusterModel(&state.StateEtcd{}, clusterState)
	plan, err := scheduler.newPlan(clusterModel, structs.ClusterFeatures{NodeCount: 3})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "WaitForAllMembers", "WaitForLeader"}
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
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell1"},
			&config.Cell{GUID: "cell2"},
			&config.Cell{GUID: "cell3"},
			&config.Cell{GUID: "cell4"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterState := structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", CellGUID: "cell1", Role: state.LeaderRole},
			&structs.Node{ID: "b", CellGUID: "cell2", Role: state.ReplicaRole},
			&structs.Node{ID: "c", CellGUID: "cell3", Role: state.ReplicaRole},
		},
	}
	clusterModel := state.NewClusterModel(&state.StateEtcd{}, clusterState)
	plan, err := scheduler.newPlan(clusterModel, structs.ClusterFeatures{NodeCount: 2})
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"RemoveRandomNode", "WaitForLeader"}
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
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell1"},
			&config.Cell{GUID: "cell2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterState := structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", CellGUID: "cell1", Role: state.LeaderRole},
			&structs.Node{ID: "b", CellGUID: "cell-unavailable", Role: state.ReplicaRole},
		},
	}
	clusterFeatures := structs.ClusterFeatures{
		NodeCount: 2,
		CellGUIDs: []string{"cell1", "cell2"},
	}
	clusterModel := state.NewClusterModel(&state.StateEtcd{}, clusterState)
	plan, err := scheduler.newPlan(clusterModel, clusterFeatures)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "WaitForAllMembers", "RemoveNode(b)", "WaitForAllMembers", "WaitForLeader"}
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
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell1"},
			&config.Cell{GUID: "cell2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterState := structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", CellGUID: "cell-unavailable", Role: state.LeaderRole},
			&structs.Node{ID: "b", CellGUID: "cell2", Role: state.ReplicaRole},
		},
	}
	clusterFeatures := structs.ClusterFeatures{
		NodeCount: 2,
		CellGUIDs: []string{"cell1", "cell2"},
	}
	clusterModel := state.NewClusterModel(&state.StateEtcd{}, clusterState)
	plan, err := scheduler.newPlan(clusterModel, clusterFeatures)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "WaitForAllMembers", "RemoveLeader(a)", "WaitForAllMembers", "WaitForLeader"}
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
		Cells: []*config.Cell{
			&config.Cell{GUID: "cell1"},
			&config.Cell{GUID: "cell2"},
		},
		Etcd: testutil.LocalEtcdConfig,
	}
	scheduler, err := NewScheduler(schedulerConfig, new(fakes.FakePatroni), logger)
	if err != nil {
		t.Fatalf("NewScheduler error: %v", err)
	}

	clusterState := structs.ClusterState{
		InstanceID: "test",
		Nodes: []*structs.Node{
			&structs.Node{ID: "a", CellGUID: "cell-x-unavailable", Role: state.LeaderRole},
			&structs.Node{ID: "b", CellGUID: "cell-y-unavailable", Role: state.ReplicaRole},
		},
	}
	clusterFeatures := structs.ClusterFeatures{
		NodeCount: 2,
		CellGUIDs: []string{"cell1", "cell2"},
	}
	clusterModel := state.NewClusterModel(&state.StateEtcd{}, clusterState)
	plan, err := scheduler.newPlan(clusterModel, clusterFeatures)
	if err != nil {
		t.Fatalf("scheduler.newPlan error: %v", err)
	}
	expectedStepTypes := []string{"AddNode", "AddNode", "WaitForAllMembers", "RemoveNode(b)", "RemoveLeader(a)", "WaitForAllMembers", "WaitForLeader"}
	stepTypes := plan.stepTypes()
	if !reflect.DeepEqual(stepTypes, expectedStepTypes) {
		t.Fatalf("plan should have steps %v, got %v", expectedStepTypes, stepTypes)
	}
}
