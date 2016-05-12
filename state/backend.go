package state

import (
	"fmt"
	"math/rand"
	"regexp"
)

// RandomReplicaNode should discover which nodes are replicas and return a random one
// FIXME - currently just picking a random node - which might be the master
func (cluster *Cluster) RandomReplicaNode() (nodeUUID string, backend string, err error) {
	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.MetaData().InstanceID)
	resp, err := cluster.etcdClient.Get(key, false, true)
	if err != nil {
		cluster.logger.Error("random-replica-node.nodes", err)
		return
	}
	item := rand.Intn(len(resp.Node.Nodes))
	nodeKey := resp.Node.Nodes[item].Key
	r, _ := regexp.Compile("/nodes/(.*)$")
	matches := r.FindStringSubmatch(nodeKey)
	nodeUUID = matches[1]

	key = fmt.Sprintf("/serviceinstances/%s/nodes/%s/backend", cluster.MetaData().InstanceID, nodeUUID)
	resp, err = cluster.etcdClient.Get(key, false, false)
	if err != nil {
		cluster.logger.Error("random-replica-node.backend", err)
		return
	}
	backend = resp.Node.Value

	return
}
