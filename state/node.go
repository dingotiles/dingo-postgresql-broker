package state

import (
	"fmt"
	"regexp"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/pivotal-golang/lager"
)

func (cluster *Cluster) AddNode(node structs.Node) (err error) {
	cluster.logger.Info("add-node", lager.Data{"node": node})
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", cluster.meta.InstanceID, node.ID)
	_, err = cluster.etcdClient.Set(key, node.BackendID, 0)
	return
}

func (cluster *Cluster) RemoveNode(node *structs.Node) error {
	cluster.logger.Info("remove-node", lager.Data{"node": node})
	key := fmt.Sprintf("/serviceinstances/%s/nodes/%s", cluster.meta.InstanceID, node.ID)
	_, err := cluster.etcdClient.Delete(key, true)
	return err
}

// if any errors, assume that cluster has no running nodes yet
func (cluster *Cluster) Nodes() (nodes []*structs.Node) {
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
		nodeID := keyRegExp.FindStringSubmatch(nodeKey)[1]
		nodes = append(nodes, &structs.Node{
			ID:        nodeID,
			BackendID: resp.Node.Value,
			PlanID:    cluster.meta.PlanID,
			ServiceID: cluster.meta.ServiceID,
		})
	}
	return nodes
}
