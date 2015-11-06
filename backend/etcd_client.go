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

// Delete deletes the given key.
//
// When recursive set to false, if the key points to a
// directory the method will fail.
//
// When recursive set to true, if the key points to a file,
// the file will be deleted; if the key points to a directory,
// then everything under the directory (including all child directories)
// will be deleted.
func (kv *EtcdClient) Delete(key string, recursive bool) (*etcd.Response, error) {
	return kv.client.Delete(fmt.Sprintf("%s%s", kv.prefix, key), recursive)
}

// Watch for a change
// If recursive is set to true the watch returns the first change under the given
// prefix since the given index.
//
// If recursive is set to false the watch returns the first change to the given key
// since the given index.
//
// To watch for the latest change, set waitIndex = 0.
//
// If a receiver channel is given, it will be a long-term watch. Watch will block at the
//channel. After someone receives the channel, it will go on to watch that
// prefix.  If a stop channel is given, the client can close long-term watch using
// the stop channel.
func (kv *EtcdClient) Watch(key string, waitIndex uint64, recursive bool,
	receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error) {
	return kv.client.Watch(fmt.Sprintf("%s%s", kv.prefix, key), waitIndex, recursive, receiver, stop)
}
