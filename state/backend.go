package state

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"

	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/utils"
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

// AllAZs lists of AZs offered by
func (cluster *Cluster) AllAZs(backends []*config.Backend) (list []string) {
	azUsage := map[string]int{}
	for _, backend := range backends {
		azUsage[backend.AvailabilityZone]++
	}
	for az := range azUsage {
		list = append(list, az)
	}
	// TEST sorting AZs for benefit of tests
	sort.Strings(list)
	return
}

// if any errors, assume that cluster has no running nodes yet
func (cluster *Cluster) usedBackendGUIDs() (backendGUIDs []string) {
	resp, err := cluster.etcdClient.Get(fmt.Sprintf("/serviceinstances/%s/nodes", cluster.MetaData().InstanceID), false, false)
	if err != nil {
		return
	}
	for _, clusterNode := range resp.Node.Nodes {
		nodeKey := clusterNode.Key
		resp, err = cluster.etcdClient.Get(fmt.Sprintf("%s/backend", nodeKey), false, false)
		if err != nil {
			cluster.logger.Error("az-used.backend", err)
			return
		}
		backendGUIDs = append(backendGUIDs, resp.Node.Value)
	}
	return
}

// backendAZsByUnusedness sorts the availability zones in order of whether this cluster is using them or not
// An AZ that is not being used at all will be early in the result.
// All known AZs are included in the result
func (cluster *Cluster) sortBackendAZsByUnusedness(backends []*config.Backend) (vs *utils.ValSorter) {
	azUsageData := map[string]int{}
	for _, az := range cluster.AllAZs(backends) {
		azUsageData[az] = 0
	}
	for _, backendGUID := range cluster.usedBackendGUIDs() {
		for _, backend := range backends {
			if backend.GUID == backendGUID {
				azUsageData[backend.AvailabilityZone]++
			}
		}
	}
	vs = utils.NewValSorter(azUsageData)
	fmt.Printf("usage %#v\n", azUsageData)
	vs.Sort()
	fmt.Printf("sorted %#v\n", vs)
	return
}

// SortedBackendsByUnusedAZs is sequence of backends to try to request new nodes for this cluster
// It prioritizes backends in availability zones that are not currently used
func (cluster *Cluster) SortedBackendsByUnusedAZs(backends []*config.Backend) []*config.Backend {
	usedBackends, unusedBackeds := cluster.usedAndUnusedBackends(backends)

	for _, az := range cluster.sortBackendAZsByUnusedness(backends).Keys {
		for _, backend := range unusedBackeds {
			if backend.AvailabilityZone == az {
				backends = append(backends, backend)
			}
		}
	}
	ret := []*config.Backend{}
	for _, backend := range usedBackends {
		ret = append(ret, backend)
	}
	return ret
}

func (cluster *Cluster) usedAndUnusedBackends(backends []*config.Backend) (usedBackends, unusuedBackends []*config.Backend) {
	usedBackendGUIDs := cluster.usedBackendGUIDs()
	for _, backend := range backends {
		used := false
		for _, usedBackendGUID := range usedBackendGUIDs {
			if backend.GUID == usedBackendGUID {
				usedBackends = append(usedBackends, backend)
				used = true
				break
			}
		}
		if !used {
			unusuedBackends = append(unusuedBackends, backend)
		}
	}
	return
}
