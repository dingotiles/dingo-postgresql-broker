package patronidata

import (
	"fmt"
	"strings"

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

// ClusterMembersRunningStates aggregates the patroni states of each member in the cluster
// allRunning is true if state of all members is "running"
func (p *Patroni) ClusterMembersRunningStates(instanceID structs.ClusterID) (statuses string, allRunning bool, err error) {
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
