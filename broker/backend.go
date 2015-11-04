package broker

import (
	"fmt"

	"github.com/coreos/go-etcd/etcd"
)

// Backend describes the location for a backend that can run Patroni nodes
type Backend struct {
	GUID     string
	URI      string
	Username string
	Password string
}

// NewBackend prepares a Backend client to which requests can be sent
func NewBackend(uri, username, password string) Backend {
	return Backend{URI: uri, Username: username, Password: password}
}

// LoadBackendsFromEtcd loads available set of backends.
// ecah +machine+ is in uri format "http://127.0.0.1:2379"
func LoadBackendsFromEtcd(machines []string, prefix string) (backend []Backend, err error) {
	client := etcd.NewClient(machines)
	resp, err := client.Get(fmt.Sprintf("%s/backends", prefix), false, true)
	if err != nil {
		return
	}
	for _, node := range resp.Node.Nodes {
		fmt.Printf("%#v\n", node)
	}
	return
}

func AddBackendToEtcd(backend Backend, machines []string, prefix string) error {
	client := etcd.NewClient(machines)
	_, err := client.Set(fmt.Sprintf("%s/backends/%s", prefix, backend.GUID), "test", 0)
	return err
}
