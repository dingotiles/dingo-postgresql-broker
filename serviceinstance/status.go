package serviceinstance

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry-community/patroni-broker/patroni"
)

// MemberStatus aggregates the patroni states of each member in the cluster
// allRunning is true if state of all members is "running"
func (cluster *Cluster) MemberStatus() (statuses string, allRunning bool) {
	key := fmt.Sprintf("/service/%s/members", cluster.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, true)
	if err != nil {
		cluster.Logger.Error("cluster.member-status.etcd-members", err)
		return fmt.Sprintf("patroni member status missing for service instance %s", cluster.InstanceID), false
	}

	masterStatus := ""
	replicasStatus := []string{}
	allRunning = true
	for _, member := range resp.Node.Nodes {
		memberData := patroni.ServiceMemberData{}
		err := json.Unmarshal([]byte(member.Value), &memberData)
		if err != nil {
			cluster.Logger.Error("cluster.member-status.etcd-member", err)
			return fmt.Sprintf("patroni member status corrupt for service instance %s", cluster.InstanceID), false
		}
		if memberData.Role == "master" {
			masterStatus = memberData.State
		} else {
			replicasStatus = append(replicasStatus, memberData.State)
		}
		if memberData.State != "running" {
			allRunning = false
		}
	}
	if masterStatus != "" {
		return fmt.Sprintf("master %s; replicas %s", masterStatus, strings.Join(replicasStatus, ", ")), allRunning
	}
	return fmt.Sprintf("members %s", strings.Join(replicasStatus, ", ")), allRunning
}
