package state

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

func resetEtcd(t *testing.T, prefix string) etcd.KeysAPI {
	if !testutil.EtcdServerAvailable(t) {
		t.SkipNow()
	}

	client, err := etcd.New(etcd.Config{Endpoints: testutil.LocalEtcdConfig.Machines})
	if err != nil {
		t.Fatalf("Failed to initialize etcd client %s", err)
		return nil
	}

	etcdApi := etcd.NewKeysAPI(client)
	_, err = etcdApi.Delete(context.Background(), prefix, &etcd.DeleteOptions{
		Recursive: true,
		Dir:       true,
	})
	if err != nil {
		t.Logf("Could not delete etcd dir %s, Error: %s", prefix, err)
	}
	return etcdApi
}

func TestState_SaveCluster(t *testing.T) {
	t.Parallel()

	testPrefix := "TestState_SaveCluster"
	etcdApi := resetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	state, err := NewStateWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create state", err)
	}

	clusterID := uuid.New()
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

	resp, err = etcdApi.Get(context.Background(), fmt.Sprintf("%s/service/%s/plan_id", testPrefix, clusterID), &etcd.GetOptions{})
	if err != nil {
		t.Fatalf("Could note get plan_id from etcd%s", err)
	}

	if want, got := planID, resp.Node.Value; want != got {
		t.Fatalf("PlanID did not match. Want %s, got %s", want, got)
	}
}
