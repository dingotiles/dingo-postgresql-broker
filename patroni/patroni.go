package patroni

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type Patroni struct {
	etcd   etcd.KeysAPI
	logger lager.Logger
}

const (
	LeaderRole  = "master"
	MasterRole  = "master"
	ReplicaRole = "replica"

	RunningState = "running"

	leaderNameDoesNotMatch = "leader name does not match"
	noOtherMembers         = "cluster does not have members except leader"
	noGoodCandidates       = "no good candidates have been found"

	waitTilMemberRunningTimeout = 300 * time.Second
	waitForLeaderTimeout        = 300 * time.Second
	failoverFromTimeout         = 300 * time.Second
)

type ClusterMember struct {
	Role         string `json:"role"`
	State        string `json:"state"`
	XlogLocation int64  `json:"xlog_location"`
	ConnURL      string `json:"conn_url"`
	APIURL       string `json:"api_url"`
	RootAPIURL   string
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

func (p *Patroni) ClusterLeader(instanceID structs.ClusterID) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("service/%s/leader", instanceID)
	resp, err := p.etcd.Get(ctx, key, &etcd.GetOptions{Quorum: true})
	if err != nil {
		p.logger.Error("patroni.cluster-leader.error", err)
		return "", err
	}
	return resp.Node.Value, nil
}

// WaitForLeader blocks until leader is elected and active
func (p *Patroni) WaitForLeader(instanceID structs.ClusterID) error {
	timeout := time.After(waitForLeaderTimeout)
	c := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for leader of %d", instanceID)
		case <-c:
			if p.leaderRunning(instanceID) {
				return nil
			}
		}
	}
	return nil
}

// WaitForAllMembers waits until expected number of nodes are running (not too many, not too few, and all running)
func (p *Patroni) WaitForAllMembers(instanceID structs.ClusterID, expectedNodeCount int) error {
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

func (p *Patroni) WaitForMember(instanceID structs.ClusterID, memberID string) error {
	notFoundRegExp, _ := regexp.Compile("Key not found")

	timeout := time.After(waitTilMemberRunningTimeout)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for member %s appear in data store", memberID)
		case <-tick:
			member, err := p.loadMember(instanceID, memberID)
			if err != nil {
				p.logger.Error("cluster-data.member-data.get", err, lager.Data{
					"instance-id":   instanceID,
					"member":        memberID,
					"err":           err.Error(),
					"not-found-yet": notFoundRegExp.MatchString(err.Error()),
				})

				if !notFoundRegExp.MatchString(err.Error()) {
					return err
				}
				p.logger.Info("cluster-data.member-data.waiting", lager.Data{
					"instance-id": instanceID,
					"member":      memberID,
				})
			} else {
				if member.State == "running" {
					return nil
				}
			}
		}
	}
	return nil
}

func (p *Patroni) FailoverFrom(instanceID structs.ClusterID, memberID string) error {
	p.logger.Info("patroni.failover-from", lager.Data{"instance-id": instanceID, "member-id": memberID})
	member, err := p.loadMember(instanceID, memberID)
	if err != nil {
		p.logger.Error("patroni.failover-from.load-member", err)
		return err
	}

	url := fmt.Sprintf("%s/failover", member.RootAPIURL)

	requestData := fmt.Sprintf("{\"leader\": \"%s\"}", memberID)
	timeout := time.After(failoverFromTimeout)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out failing over %s from", instanceID, memberID)
		case <-tick:
			req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestData))
			if err != nil {
				p.logger.Error("patroni.failover-from.failover-req", err)
				return err
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				p.logger.Error("patroni.failover-from.failover-resp", err)
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				return nil
			}

			responseText, _ := ioutil.ReadAll(resp.Body)
			if match, _ := regexp.Match(leaderNameDoesNotMatch, responseText); match == true {
				p.logger.Info("patroni.failover-from.leader-does-not-match", lager.Data{"instance-id": instanceID, "member-id": memberID})
				return nil
			}
			if match, _ := regexp.Match(noOtherMembers, responseText); match == true {
				p.logger.Info("patroni.failover-from.no-other-members", lager.Data{"instance-id": instanceID, "leader-id": memberID})
				return nil
			}
			if match, _ := regexp.Match(noGoodCandidates, responseText); match == true {
				p.logger.Info("patroni.failover-from.no-good-candidates", lager.Data{"instance-id": instanceID, "leader-id": memberID})
				continue
			}

			return fmt.Errorf("Unknown error: '%s'", string(responseText))
		}
	}
}

// TODO: prove list of member IDs that cannot be member OR that can be member
// This will ensure that success isn't for an ex-leader that hasn't died yet
func (p *Patroni) leaderRunning(instanceID structs.ClusterID) bool {
	p.logger.Info("check-leader.find-leader", lager.Data{"instanceID": instanceID})
	leaderID, err := p.ClusterLeader(instanceID)
	if err != nil {
		return false
	}
	p.logger.Info("check-leader.load-member", lager.Data{"instanceID": instanceID, "leader": leaderID})
	leader, err := p.loadMember(instanceID, leaderID)
	if err != nil {
		return false
	}
	p.logger.Info("check-leader.leader", lager.Data{"leader": leaderID, "data": leader})
	return leader.State == RunningState && leader.Role == LeaderRole
}

// Checks that the expected number of nodes are running (not too many, not too few, and all running)
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
		member, err := p.deserializeMember(member.Value)
		if err != nil {
			p.logger.Error("members-data.etcd-members.decode", err)
			return false
		}
		if member.State != "running" {
			return false
		}
	}
	return true
}

func setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}

func (p *Patroni) loadMember(instanceID structs.ClusterID, memberID string) (member ClusterMember, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("service/%s/members/%s", instanceID, memberID)
	resp, err := p.etcd.Get(ctx, key, &etcd.GetOptions{Quorum: true})
	if err != nil {
		p.logger.Error("load-member.etcd-get", err, lager.Data{"member": memberID, "key": key})
		return
	}
	member, err = p.deserializeMember(resp.Node.Value)
	if err != nil {
		p.logger.Error("load-member.decode", err, lager.Data{"member": memberID})
		return
	}
	return
}

func (p *Patroni) deserializeMember(jsonValue string) (member ClusterMember, err error) {
	dec := json.NewDecoder(strings.NewReader(jsonValue))
	if err = dec.Decode(&member); err == io.EOF {
		return
	} else if err != nil {
		return
	}

	member.RootAPIURL = strings.Replace(member.APIURL, "/patroni", "/", 1)

	return
}
