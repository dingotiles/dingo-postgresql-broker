package routing

import (
	"sort"
	"testing"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
	"github.com/pivotal-golang/lager"
)

func resetEtcd(t *testing.T, prefix string) {
	if !testutil.EtcdServerAvailable(t) {
		t.SkipNow()
	}

	client, err := etcd.New(etcd.Config{Endpoints: testutil.LocalEtcdConfig.Machines})
	if err != nil {
		t.Fatalf("Failed to initialize etcd client %s", err)
		return
	}

	api := etcd.NewKeysAPI(client)
	_, err = api.Delete(context.Background(), prefix, &etcd.DeleteOptions{
		Recursive: true,
		Dir:       true,
	})
	if err != nil {
		t.Logf("Could not delete etcd dir %s, Error: %s", prefix, err)
	}
	return
}

type logAdapter struct {
	t *testing.T
}

func (l logAdapter) Log(level lager.LogLevel, msg []byte) {
	l.t.Logf("Logged message: %s", string(msg))
}
func testLogger(t *testing.T) lager.Logger {
	logger := lager.NewLogger("router-test")
	logger.RegisterSink(logAdapter{t: t})
	return logger
}

func TestRouter_InitialPort(t *testing.T) {
	t.Parallel()

	testPrefix := "TestRouter_InitialPort"
	resetEtcd(t, testPrefix)

	router, err := NewRouterWithPrefix(testutil.LocalEtcdConfig, testPrefix, testLogger(t))
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

func TestRouter_ConcurrentPortAllocation(t *testing.T) {
	t.Parallel()

	testPrefix := "TestRouter_ConcurrentPortAllocation"
	resetEtcd(t, testPrefix)

	router, err := NewRouterWithPrefix(testutil.LocalEtcdConfig, testPrefix, testLogger(t))
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
