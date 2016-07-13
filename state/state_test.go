package state

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

func TestState_SaveCluster(t *testing.T) {
	t.Parallel()

	testPrefix := "TestState_SaveCluster"
	etcdApi := testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	state, err := NewStateEtcdWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create state", err)
	}

	clusterID := structs.ClusterID(uuid.New())
	planID := uuid.New()
	clusterState := structs.ClusterState{
		InstanceID:       clusterID,
		OrganizationGUID: "OrganizationGUID",
		PlanID:           planID,
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
	}
	err = state.SaveCluster(clusterState)
	if err != nil {
		t.Fatalf("SaveCluster failed %s", err)
	}

	resp, err := etcdApi.Get(context.Background(), fmt.Sprintf("%s/service/%s/state", testPrefix, clusterID), &etcd.GetOptions{})
	if err != nil {
		t.Fatalf("Could not load state from etcd %s", err)
	}

	retrievedState := structs.ClusterState{}
	json.Unmarshal([]byte(resp.Node.Value), &retrievedState)
	if !reflect.DeepEqual(clusterState, retrievedState) {
		t.Fatalf("Retrieved State does not match. Want %v, Got %v", clusterState, retrievedState)
	}
}

func TestState_ClusterExists(t *testing.T) {
	t.Parallel()

	testPrefix := "TestState_ClusterExists"
	testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	state, err := NewStateEtcdWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create state", err)
	}

	clusterID := structs.ClusterID(uuid.New())
	planID := uuid.New()
	clusterState := structs.ClusterState{
		InstanceID:       clusterID,
		OrganizationGUID: "OrganizationGUID",
		PlanID:           planID,
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
	}
	err = state.SaveCluster(clusterState)
	if err != nil {
		t.Fatalf("SaveCluster failed %s", err)
	}

	if !state.ClusterExists(clusterID) {
		t.Fatalf("Cluster %s should exist", clusterID)
	}

	if state.ClusterExists("fakeID") {
		t.Fatalf("Cluster %s should not exist", "fakeID")
	}
}

func TestState_LoadCluster(t *testing.T) {
	t.Parallel()

	testPrefix := "TestState_LoadClusterState"
	testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	state, err := NewStateEtcdWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create state", err)
	}

	instanceID := structs.ClusterID(uuid.New())
	planID := uuid.New()
	clusterState := structs.ClusterState{
		InstanceID:       instanceID,
		OrganizationGUID: "OrganizationGUID",
		PlanID:           planID,
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
	}
	err = state.SaveCluster(clusterState)
	if err != nil {
		t.Fatalf("SaveCluster failed %s", err)
	}

	loadedState, err := state.LoadCluster(instanceID)
	if !reflect.DeepEqual(clusterState, loadedState) {
		t.Fatalf("Failed to load ClusterState")
	}
}

func TestState_DeleteCluster(t *testing.T) {
	t.Parallel()

	testPrefix := "TestState_DeleteClusterState"
	etcdApi := testutil.ResetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	state, err := NewStateEtcdWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create state", err)
	}

	instanceID := structs.ClusterID(uuid.New())
	planID := uuid.New()
	clusterState := structs.ClusterState{
		InstanceID:       instanceID,
		OrganizationGUID: "OrganizationGUID",
		PlanID:           planID,
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
	}
	err = state.SaveCluster(clusterState)
	if err != nil {
		t.Fatalf("SaveCluster failed %s", err)
	}

	err = state.DeleteCluster(instanceID)
	if err != nil {
		t.Fatalf("DeleteClusterState failed %s", err)
	}

	key := fmt.Sprintf("%s/service/%s/state", testPrefix, instanceID)
	_, err = etcdApi.Get(context.Background(), key, &etcd.GetOptions{})
	if err == nil {
		t.Fatalf("Was expecting error 'Key not found'")
	} else {
		notFoundRegExp, _ := regexp.Compile("Key not found")
		if notFoundRegExp.FindString(err.Error()) != "Key not found" {
			t.Fatalf("An error other than 'Key not found' occured %s", err)
		}
	}
}
