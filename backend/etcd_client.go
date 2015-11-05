package backend

import (
	"fmt"

	"github.com/coreos/go-etcd/etcd"
)

// EtcdClient is a wrapper for an etcd.Client; and common KV functions
type EtcdClient struct {
	client *etcd.Client
	prefix string
}

// NewEtcdClient creates an *EtcdClient
func NewEtcdClient(machines []string, prefix string) (kv *EtcdClient) {
	kv = &EtcdClient{client: etcd.NewClient(machines), prefix: prefix}
	return
}

// Get gets the file or directory associated with the given key.
// If the key points to a directory, files and directories under
// it will be returned in sorted or unsorted order, depending on
// the sort flag.
// If recursive is set to false, contents under child directories
// will not be returned.
// If recursive is set to true, all the contents will be returned.
func (kv *EtcdClient) Get(key string, sort, recursive bool) (*etcd.Response, error) {
	return kv.client.Get(fmt.Sprintf("%s%s", kv.prefix, key), sort, recursive)
}

// Set sets the given key to the given value.
// It will create a new key value pair or replace the old one.
// It will not replace a existing directory.
func (kv *EtcdClient) Set(key string, value string, ttl uint64) (*etcd.Response, error) {
	return kv.client.Set(fmt.Sprintf("%s%s", kv.prefix, key), value, ttl)
}
