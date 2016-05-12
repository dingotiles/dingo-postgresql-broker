package state

import (
	"fmt"
	"regexp"
)

type Node struct {
	Id        string
	BackendId string
	PlanId    string
	ServiceId string
}

func (cluster *Cluster) AddNode(node Node) (err error) {
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", cluster.meta.InstanceID, node.Id)
	_, err = cluster.etcdClient.Set(key, node.BackendId, 0)
	return
}

func (cluster *Cluster) RemoveNode(node *Node) error {
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s", cluster.meta.InstanceID, node.Id)
	_, err := cluster.etcdClient.Delete(key, true)
	return err
}

// if any errors, assume that cluster has no running nodes yet
func (cluster *Cluster) Nodes() (nodes []*Node) {
	resp, err := cluster.etcdClient.Get(fmt.Sprintf("/serviceinstances/%s/nodes", cluster.MetaData().InstanceID), false, false)
	if err != nil {
		return
	}

	keyRegExp, _ := regexp.Compile("/nodes/(.*)$")
	for _, clusterNode := range resp.Node.Nodes {
		nodeKey := clusterNode.Key
		resp, err = cluster.etcdClient.Get(fmt.Sprintf("%s/backend", nodeKey), false, false)
		if err != nil {
			cluster.logger.Error("az-used.backend", err)
			return
		}
		nodeId := keyRegExp.FindStringSubmatch(nodeKey)[1]
		nodes = append(nodes, &Node{
			Id:        nodeId,
			BackendId: resp.Node.Value,
			PlanId:    cluster.meta.PlanID,
			ServiceId: cluster.meta.ServiceID,
		})
	}
	return nodes
}
