package backend

import "github.com/coreos/go-etcd/etcd"

// FakeEtcdClient is a fake etcd client for test purposes.
type FakeEtcdClient struct {
	GetResponses map[string]*etcd.Response
}

// NewFakeEtcdClient creates an *FakeEtcdClient
func NewFakeEtcdClient() (kv FakeEtcdClient) {
	kv.GetResponses = map[string]*etcd.Response{}
	return
}

// Get gets the file or directory associated with the given key.
// If the key points to a directory, files and directories under
// it will be returned in sorted or unsorted order, depending on
// the sort flag.
// If recursive is set to false, contents under child directories
// will not be returned.
// If recursive is set to true, all the contents will be returned.
func (kv FakeEtcdClient) Get(key string, sort, recursive bool) (resp *etcd.Response, err error) {
	if kv.GetResponses[key] != nil {
		return kv.GetResponses[key], nil
	}
	emptyResp := &etcd.Response{
		Node: &etcd.Node{},
	}
	return emptyResp, nil
	// resp = &etcd.Response{
	// 	Node: &etcd.Node{
	// 		Nodes: etcd.Nodes{
	// 			&etcd.Node{Key: "1", Value: "{}"},
	// 			&etcd.Node{Key: "1", Value: "{}"},
	// 		},
	// 	},
	// }
}

// Set sets the given key to the given value.
// It will create a new key value pair or replace the old one.
// It will not replace a existing directory.
func (kv FakeEtcdClient) Set(key string, value string, ttl uint64) (resp *etcd.Response, err error) {
	resp = &etcd.Response{}
	return
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
func (kv FakeEtcdClient) Delete(key string, recursive bool) (resp *etcd.Response, err error) {
	resp = &etcd.Response{}
	return
}
