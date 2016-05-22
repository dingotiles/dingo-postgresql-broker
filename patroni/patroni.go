package patroni

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

// ServiceMemberData contains the data advertised by a patroni member
type ServiceMemberData struct {
	ConnURL  string `json:"conn_url"`
	HostPort string `json:"conn_address"`
	Role     string `json:"role"`
	State    string `json:"state"`
}

// MemberStatus aggregates the patroni states of each member in the cluster
// allRunning is true if state of all members is "running"
func MemberStatus(instanceID string, etcdConf config.Etcd, logger lager.Logger) (statuses string, allRunning bool, err error) {
	etcdClient, err := setupEtcd(etcdConf)
	if err != nil {
		return "", false, err
	}

	ctx := context.Background()
	key := fmt.Sprintf("/service/%s/members", instanceID)
	resp, err := etcdClient.Get(ctx, key, &etcd.GetOptions{
		Quorum:    true,
		Recursive: true,
	})
	if err != nil {
		logger.Error("member-status.etcd-members", err)
		return fmt.Sprintf("patroni member status missing for service instance %s", instanceID), false, err
	}

	masterStatus := ""
	replicasStatus := []string{}
	allRunning = true
	for _, member := range resp.Node.Nodes {
		memberData := ServiceMemberData{}
		err := json.Unmarshal([]byte(member.Value), &memberData)
		if err != nil {
			logger.Error("member-status.etcd-member", err)
			return fmt.Sprintf("patroni member status corrupt for service instance %s", instanceID), false, err
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

func setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}