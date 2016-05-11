package state

import "fmt"

type Node struct {
	Id        string
	BackendId string
}

func (cluster *Cluster) AddNode(node Node) (kvIndex uint64, err error) {
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", cluster.meta.InstanceID, node.Id)
	resp, err := cluster.etcdClient.Set(key, node.BackendId, 0)
	if err != nil {
		return 0, err
	}
	return resp.EtcdIndex, err
}

func (cluster *Cluster) RemoveNode(nodeId string) error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s", cluster.meta.InstanceID, nodeId)
	_, err := cluster.etcdClient.Delete(key, true)
	return err
}
