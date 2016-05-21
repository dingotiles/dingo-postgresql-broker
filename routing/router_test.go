package routing

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
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

func TestRouter_InitialPort(t *testing.T) {
	t.Parallel()

	testPrefix := "TestRouter_InitialPort"
	resetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	router, err := NewRouterWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatal("Could not create a new router", err)
	}

	nextPort, err := router.AllocatePort()
	if err != nil {
		t.Fatal("Could not allocate port", err)
	}

	if nextPort != initialPort {
		t.Errorf("%s was not initialized in etcd", nextPortKey)
	}
}

func TestRouter_WithoutPrefix(t *testing.T) {
	t.Parallel()

	testPrefix := ""
	resetEtcd(t, "routing")
	logger := testutil.NewTestLogger(testPrefix, t)

	router, err := NewRouter(testutil.LocalEtcdConfig, logger)
	if err != nil {
		t.Fatal("Could not create a new router", err)
	}

	nextPort, err := router.AllocatePort()
	if err != nil {
		t.Fatal("Could not allocate port", err)
	}

	if nextPort != initialPort {
		t.Errorf("Allocated port did not equal initial port. Want %d, got %d", initialPort, nextPort)
	}
}

func TestRouter_ConcurrentPortAllocation(t *testing.T) {
	t.Parallel()

	testPrefix := "TestRouter_ConcurrentPortAllocation"
	resetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	router, err := NewRouterWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create a new router", err)
	}

	portChan := make(chan int)
	for i := 0; i < 5; i++ {
		go func() {
			nextPort, err := router.AllocatePort()
			if err != nil {
				portChan <- 0
				t.Error("Could not allocate port", err)
			}
			portChan <- nextPort
		}()
	}

	ports := []int{}
	for i := 0; i < 5; i++ {
		ports = append(ports, <-portChan)
	}

	sort.Ints(ports)
	for i := 0; i < 5; i++ {
		if want, got := initialPort+i, ports[i]; want != got {
			t.Errorf("Concurrent allocation of ports failed. Expected %d, got %d", want, got)
		}
	}
}

func TestRouter_AssignPortToCluster(t *testing.T) {
	t.Parallel()

	testPrefix := "TestRouter_AssignPortToCluster"
	etcdApi := resetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	router, err := NewRouterWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create a new router", err)
	}

	clusterID := "clusterID"
	port := 30000
	err = router.AssignPortToCluster(clusterID, port)
	if err != nil {
		t.Fatalf("Assigning port failed")
	}

	key := fmt.Sprintf("%s/routing/allocation/%s", testPrefix, clusterID)
	resp, err := etcdApi.Get(context.Background(), key, &etcd.GetOptions{})
	if err != nil {
		t.Fatalf("Could not read port from etcd")
	}

	retrievedPort, err := strconv.Atoi(resp.Node.Value)
	if want, got := port, retrievedPort; want != got {
		t.Fatalf("Routing was not initialized. Expected %d, got %d", want, got)
	}
}

func TestRouter_RemoveClusterAssignement(t *testing.T) {
	t.Parallel()

	testPrefix := "TestRouter_RemoveClusterAssignement"
	etcdApi := resetEtcd(t, testPrefix)
	logger := testutil.NewTestLogger(testPrefix, t)

	clusterID := "clusterID"
	port := 30000

	key := fmt.Sprintf("%s/routing/allocation/%s", testPrefix, clusterID)
	_, err := etcdApi.Set(context.Background(), key, fmt.Sprintf("%d", port), &etcd.SetOptions{})

	router, err := NewRouterWithPrefix(testutil.LocalEtcdConfig, testPrefix, logger)
	if err != nil {
		t.Fatalf("Could not create a new router", err)
	}

	err = router.RemoveClusterAssignment(clusterID)
	if err != nil {
		t.Fatalf("Could not remove the assignment %s", err)
	}

	_, err = etcdApi.Get(context.Background(), fmt.Sprintf("%s/routing/allocation/%s", testPrefix, clusterID), &etcd.GetOptions{})
	if err == nil {
		t.Fatalf("port wasn't deleted %s", err)
	}
}
