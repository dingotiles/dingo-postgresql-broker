package testutil

import (
	"testing"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"golang.org/x/net/context"
)

var (
	LocalEtcdConfig = config.Etcd{Machines: []string{"http://localhost:4001"}}
)

func EtcdServerAvailable(t *testing.T) bool {
	config := etcd.Config{
		Endpoints: LocalEtcdConfig.Machines,
	}
	client, err := etcd.New(config)
	if err != nil {
		t.Fatalf("Failed to initialize etcd client %s", err)
		return false
	}

	api := etcd.NewKeysAPI(client)
	_, err = api.Get(context.Background(), "", &etcd.GetOptions{})
	if err != nil {
		t.Logf("Failed to connect to etcd %s", err)
		return false
	}
	t.Logf("Etcd client is available on localhost")

	return true
}
