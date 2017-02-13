package structs

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	defaultNodeCount = 1

	// SchedulingStatusUnknown is a sad state of affairs
	SchedulingStatusUnknown = SchedulingStatus("")
	// SchedulingStatusSuccess means a scheduled change to a cluster has concluded successfully
	SchedulingStatusSuccess = SchedulingStatus("success")
	// SchedulingStatusInProgress means a scheduled change is in-progress
	SchedulingStatusInProgress = SchedulingStatus("in-progress")
	// SchedulingStatusFailed means a scheduled change has finished but failed
	SchedulingStatusFailed = SchedulingStatus("failed")
)

// SchedulingStatus is a set of known states for a schedule
type SchedulingStatus string

// ClusterID aka InstanceID is the unique ID for a Cluster
type ClusterID string

// ClusterRecreationData is the stored state of a Cluster for future recreation/restoration
type ClusterRecreationData struct {
	ServiceInstanceName  string              `json:"service_instance_name"`
	InstanceID           ClusterID           `json:"instance_id"`
	ServiceID            string              `json:"service_id"`
	PlanID               string              `json:"plan_id"`
	OrganizationGUID     string              `json:"organization_guid"`
	SpaceGUID            string              `json:"space_guid"`
	AdminCredentials     PostgresCredentials `json:"admin_credentials"`
	SuperuserCredentials PostgresCredentials `json:"superuser_credentials"`
	AppCredentials       PostgresCredentials `json:"app_credentials"`
	AllocatedPort        int                 `json:"allocated_port"`
}

// ClusterState documents known state of a Cluster
type ClusterState struct {
	InstanceID           ClusterID           `json:"instance_id"`
	ServiceID            string              `json:"service_id"`
	PlanID               string              `json:"plan_id"`
	OrganizationGUID     string              `json:"organization_guid"`
	SpaceGUID            string              `json:"space_guid"`
	AdminCredentials     PostgresCredentials `json:"admin_credentials"`
	SuperuserCredentials PostgresCredentials `json:"superuser_credentials"`
	AppCredentials       PostgresCredentials `json:"app_credentials"`
	AllocatedPort        int                 `json:"allocated_port"`
	SchedulingInfo       SchedulingInfo      `json:"info"`
	ServiceInstanceName  string              `json:"service_instance_name"`
	Nodes                []*Node             `json:"nodes"`
}

// SchedulingInfo is that status of an in-progress scheduled change to a cluster
type SchedulingInfo struct {
	Status         SchedulingStatus `json:"status"`
	Steps          int              `json:"steps"`
	CompletedSteps int              `json:"completed_steps"`
	LastMessage    string           `json:"last_message"`
}

// NodeCount is the number of nodes running in this cluster
func (c *ClusterState) NodeCount() int {
	return len(c.Nodes)
}

// RecreationData is a subset of ClusterState that is remotely stored
// to allow a cluster to be re-built/resurrected/restored in future
// It is all the known promises to end users and internal secrets.
func (c *ClusterState) RecreationData() *ClusterRecreationData {
	return &ClusterRecreationData{
		ServiceInstanceName:  c.ServiceInstanceName,
		InstanceID:           c.InstanceID,
		ServiceID:            c.ServiceID,
		PlanID:               c.PlanID,
		OrganizationGUID:     c.OrganizationGUID,
		SpaceGUID:            c.SpaceGUID,
		AdminCredentials:     c.AdminCredentials,
		SuperuserCredentials: c.SuperuserCredentials,
		AppCredentials:       c.AppCredentials,
		AllocatedPort:        c.AllocatedPort,
	}
}

// AddNode updates the known state of a cluster
func (c *ClusterState) AddNode(node Node) {
	c.Nodes = append(c.Nodes, &node)
}

// RemoveNode updates the known state of a cluster
func (c *ClusterState) RemoveNode(node *Node) {
	for i, n := range c.Nodes {
		if n.ID == node.ID {
			c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
			break
		}
	}
}

// NodeOnCell returns the Node that was place on a particular cell
func (c *ClusterState) NodeOnCell(cellGUID string) (Node, error) {
	for _, node := range c.Nodes {
		if node.CellGUID == cellGUID {
			return *node, nil
		}
	}
	return Node{}, fmt.Errorf("Cluster %s has no node on cell %s", c.InstanceID, cellGUID)
}

// ClusterFeatures describes a desired state for a cluster
type ClusterFeatures struct {
	NodeCount            int      `mapstructure:"node-count"`
	CellGUIDs            []string `mapstructure:"cells"`
	CloneFromServiceName string   `mapstructure:"clone-from"`
}

// PostgresCredentials describes basic auth credentials for PostgreSQL
type PostgresCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Node represents the allocation of a node to a cell within a cluster
type Node struct {
	ID       string `json:"node_id"`
	CellGUID string `json:"cell_guid"`
}

// ClusterFeaturesFromParameters customizes changes to a cluster from parameters from user
func ClusterFeaturesFromParameters(params map[string]interface{}) (features ClusterFeatures, err error) {
	err = mapstructure.Decode(params, &features)
	if err != nil {
		return
	}
	if features.NodeCount == 0 {
		features.NodeCount = defaultNodeCount
	}
	if features.NodeCount < 0 {
		err = fmt.Errorf("Broker: node-count (%d) must be a positive number", features.NodeCount)
		return
	}

	return
}
