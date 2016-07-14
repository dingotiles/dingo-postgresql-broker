package patronidata

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/pivotal-golang/lager"
)

const (
	waitTilMemberRunningTimeout = 300 * time.Second
	waitForLeaderTimeout        = 300 * time.Second
)

// ClusterDataWrapper allows access to latest cluster information for a specific cluster
type ClusterDataWrapperReal struct {
	patroni    *Patroni
	instanceID structs.ClusterID
}

type ClusterDataWrapper interface {
	WaitTilMemberExists(memberID string) error
}

// NewClusterDataWrapper creates a ClusterDataWrapper
func NewClusterDataWrapper(patroni *Patroni, instanceID structs.ClusterID) ClusterDataWrapper {
	return ClusterDataWrapperReal{
		patroni:    patroni,
		instanceID: instanceID,
	}
}

// WaitTilMemberExists blocks until cluster member exists in data store
func (cluster ClusterDataWrapperReal) WaitTilMemberExists(memberID string) error {
	notFoundRegExp, _ := regexp.Compile("Key not found")

	timeout := time.After(waitTilMemberRunningTimeout)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for member %s appear in data store", memberID)
		case <-tick:
			member, err := cluster.patroni.LoadMember(cluster.instanceID, memberID)
			if err != nil {
				cluster.patroni.logger.Error("cluster-data.member-data.get", err, lager.Data{
					"instance-id":   cluster.instanceID,
					"member":        memberID,
					"err":           err.Error(),
					"not-found-yet": notFoundRegExp.MatchString(err.Error()),
				})

				if !notFoundRegExp.MatchString(err.Error()) {
					return err
				}
				cluster.patroni.logger.Info("cluster-data.member-data.waiting", lager.Data{
					"instance-id": cluster.instanceID,
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
