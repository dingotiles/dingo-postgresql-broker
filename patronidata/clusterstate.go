package patronidata

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
)

const (
	waitTilMemberRunningTimeout = 300 * time.Second
)

// ClusterDataWrapper allows access to latest cluster information for a specific cluster
type ClusterDataWrapperReal struct {
	patroni    *Patroni
	instanceID structs.ClusterID
}

type ClusterDataWrapper interface {
	WaitTilMemberExists(memberID string) error
	WaitTilMemberRunning(memberID string) error
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
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(waitTilMemberRunningTimeout)
		timeout <- true
	}()

	c := time.Tick(1 * time.Second)
	for {
		select {
		case <-c:
			memberData, err := cluster.patroni.MemberData(cluster.instanceID, memberID)
			if err != nil {
				notFoundRegExp, _ := regexp.Compile("Key not found")
				if notFoundRegExp.FindString(err.Error()) != "Key not found" {
					return err
				}
			}
			if memberData.State == "running" {
				return nil
			}
		case <-timeout:
			return fmt.Errorf("Timed out waiting for member %s appear in data store", memberID)
		}
	}
	return nil
}

// WaitTilMemberRunning blocks until cluster member's state is "running"
func (cluster ClusterDataWrapperReal) WaitTilMemberRunning(memberID string) error {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(waitTilMemberRunningTimeout)
		timeout <- true
	}()

	c := time.Tick(1 * time.Second)
	for {
		select {
		case <-c:
			memberData, err := cluster.patroni.MemberData(cluster.instanceID, memberID)
			if err != nil {
				return err
			}
			if memberData.State == "running" {
				return nil
			}
		case <-timeout:
			return fmt.Errorf("Timed out waiting for member %s to achieve state 'running'", memberID)
		}
	}
	return nil
}
