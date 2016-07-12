package patronidata

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/patroniclient/datastructs"
	"github.com/pivotal-golang/lager"
)

type Patroni struct {
	etcd   etcd.KeysAPI
	logger lager.Logger
}

func NewPatroni(etcdConf config.Etcd, logger lager.Logger) (*Patroni, error) {
	etcd, err := setupEtcd(etcdConf)
	if err != nil {
		return nil, err
	}

	return &Patroni{
		etcd:   etcd,
		logger: logger,
	}, nil
}

// LoadCluster fetches the latest data/state for specific cluster
func (p *Patroni) MemberData(instanceID structs.ClusterID, memberID string) (memberData *datastructs.DataServiceMember, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("service/%s/members/%s", instanceID, memberID)
	resp, err := p.etcd.Get(ctx, key, &etcd.GetOptions{Quorum: true})
	if err != nil {
		p.logger.Error("cluster-data.member-data.etcd-get", err, lager.Data{"member": memberID, "key": key})
		return
	}
	memberData, err = datastructs.NewDataServiceMember(resp.Node.Value)
	if err != nil {
		p.logger.Error("cluster-data.member-data.decode", err, lager.Data{"member": memberID})
		return
	}
	return
}

// WaitTilMemberRunning blocks until every cluster member's state is "running"
func (p *Patroni) WaitTilClusterMembersRunning(instanceID structs.ClusterID, expectedNodeCount int) error {
	timeout := time.After(waitTilMemberRunningTimeout)
	c := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for cluster %d members to achieve state 'running'", instanceID)
		case <-c:
			if p.checkClusterMembersRunning(instanceID, expectedNodeCount) {
				return nil
			}
		}
	}
	return nil
}

func (p *Patroni) checkClusterMembersRunning(instanceID structs.ClusterID, expectedNodeCount int) bool {
	var err error

	// Right number of nodes?
	ctx := context.Background()
	key := fmt.Sprintf("service/%s/members", instanceID)
	resp, err := p.etcd.Get(ctx, key, &etcd.GetOptions{
		Quorum:    true,
		Recursive: true,
	})
	if err != nil {
		p.logger.Error("members-data.etcd-members.fetch", err, lager.Data{"instance-id": instanceID})
		return false
	}
	missingNodes := expectedNodeCount - len(resp.Node.Nodes)
	if missingNodes != 0 {
		return false
	}

	for _, member := range resp.Node.Nodes {
		memberData, err := datastructs.NewDataServiceMember(member.Value)
		if err != nil {
			p.logger.Error("members-data.etcd-members.decode", err)
			return false
		}
		if memberData.State != "running" {
			return false
		}
	}
	return true
}

// ClusterMembersRunningStates aggregates the patroni states of each member in the cluster
// allRunning is true if state of all members is "running"
func (p *Patroni) ClusterMembersRunningStates(instanceID structs.ClusterID, expectedNodeCount int) (statuses string, allRunning bool, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("service/%s/members", instanceID)
	resp, err := p.etcd.Get(ctx, key, &etcd.GetOptions{
		Quorum:    true,
		Recursive: true,
	})
	if err != nil {
		p.logger.Error("member-status.etcd-members", err)
		return fmt.Sprintf("patroni member status missing for service instance %s", instanceID), false, err
	}

	masterStatus := ""
	replicasStatus := []string{}
	allRunning = true
	for _, member := range resp.Node.Nodes {
		memberData, err := datastructs.NewDataServiceMember(member.Value)
		if err != nil {
			p.logger.Error("member-status.etcd-member", err)
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
	missingNodes := expectedNodeCount - len(resp.Node.Nodes)
	if missingNodes > 0 {
		for i := 0; i < missingNodes; i++ {
			replicasStatus = append(replicasStatus, "missing")
		}
	}
	// if there are currently too many or too few nodes, then cluster cannot be "all running"
	if missingNodes != 0 {
		allRunning = false
	}

	if masterStatus != "" {
		return fmt.Sprintf("master %s; replicas %s", masterStatus, strings.Join(replicasStatus, ", ")), allRunning, nil
	}
	return fmt.Sprintf("members %s", strings.Join(replicasStatus, ", ")), allRunning, nil
}

func (p *Patroni) ClusterLeader(instanceID structs.ClusterID) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("service/%s/leader", instanceID)
	resp, err := p.etcd.Get(ctx, key, &etcd.GetOptions{})
	if err != nil {
		p.logger.Error("patroni.cluster-leader.error", err)
		return "", err
	}
	return resp.Node.Value, nil
}

func setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}
