package backend

import "fmt"

// Backend describes the location for a backend that can run Patroni nodes
type Backend struct {
	GUID             string
	AvailabilityZone string
	URI              string
	Username         string
	Password         string
}

// NewBackend prepares a Backend client to which requests can be sent
func NewBackend() Backend {
	return Backend{AvailabilityZone: "z1"}
}

// LoadBackendsFromEtcd loads available set of backends.
// ecah +machine+ is in uri format "http://127.0.0.1:2379"
func LoadBackendsFromEtcd(etcdClient *EtcdClient) (backend []Backend, err error) {
	resp, err := etcdClient.Get("/backends", false, true)
	if err != nil {
		return
	}
	for _, node := range resp.Node.Nodes {
		fmt.Printf("%#v\n", node)
	}
	return
}

func AddBackendToEtcd(etcdClient *EtcdClient, backend Backend) error {
	_, err := etcdClient.Set(fmt.Sprintf("/backends/%s", backend.GUID), "test", 0)
	return err
}
