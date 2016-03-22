package serviceinstance

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dingotiles/patroni-broker/patroni"
	"github.com/pivotal-golang/lager"
)

// WaitForAllRunning blocks until all cluster members have state "running"
func (cluster *Cluster) WaitForAllRunning() (err error) {
	waitTimeout := 120
	waitTime := 0
	status := ""
	cluster.Logger.Debug("member-status.waiting-for-all-running.start", lager.Data{"waiting": waitTimeout})
	allRunning := false
	for ; !allRunning && waitTime < waitTimeout; waitTime++ {
		status, allRunning, err = cluster.MemberStatus()
		time.Sleep(1 * time.Second)
	}
	cluster.Logger.Debug("member-status.waiting-for-all-running.finish", lager.Data{
		"wait-time":      waitTime,
		"cluster-status": status,
		"error":          err,
	})
	cluster.Load()
	return err
}

// MemberStatus aggregates the patroni states of each member in the cluster
// allRunning is true if state of all members is "running"
func (cluster *Cluster) MemberStatus() (statuses string, allRunning bool, err error) {
	key := fmt.Sprintf("/service/%s/members", cluster.Data.InstanceID)
	resp, err := cluster.EtcdClient.Get(key, false, true)
	if err != nil {
		cluster.Logger.Error("member-status.etcd-members", err)
		return fmt.Sprintf("patroni member status missing for service instance %s", cluster.Data.InstanceID), false, err
	}

	masterStatus := ""
	replicasStatus := []string{}
	allRunning = true
	for _, member := range resp.Node.Nodes {
		memberData := patroni.ServiceMemberData{}
		err := json.Unmarshal([]byte(member.Value), &memberData)
		if err != nil {
			cluster.Logger.Error("member-status.etcd-member", err)
			return fmt.Sprintf("patroni member status corrupt for service instance %s", cluster.Data.InstanceID), false, err
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
		return fmt.Sprintf("master %s; replicas %s", masterStatus, strings.Join(replicasStatus, ", ")), allRunning, nil
	}
	return fmt.Sprintf("members %s", strings.Join(replicasStatus, ", ")), allRunning, nil
}
