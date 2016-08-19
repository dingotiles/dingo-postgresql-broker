package structs

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	defaultNodeCount = 2

	SchedulingStatusUnknown    = SchedulingStatus("")
	SchedulingStatusSuccess    = SchedulingStatus("success")
	SchedulingStatusInProgress = SchedulingStatus("in-progress")
	SchedulingStatusFailed     = SchedulingStatus("failed")
)

type SchedulingStatus string
type ClusterID string

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

type SchedulingInfo struct {
	Status         SchedulingStatus `json:"status"`
	Steps          int              `json:"steps"`
	CompletedSteps int              `json:"completed_steps"`
	LastMessage    string           `json:"last_message"`
}

func (c *ClusterState) NodeCount() int {
	return len(c.Nodes)
}

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

func (c *ClusterState) AddNode(node Node) {
	c.Nodes = append(c.Nodes, &node)
}

func (c *ClusterState) RemoveNode(node *Node) {
	for i, n := range c.Nodes {
		if n.ID == node.ID {
			c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
			break
		}
	}
}
func (m *ClusterState) NodeOnCell(cellGUID string) (Node, error) {
	for _, node := range m.Nodes {
		if node.CellGUID == cellGUID {
			return *node, nil
		}
	}
	return Node{}, fmt.Errorf("Cluster %s has no node on cell %s", m.InstanceID, cellGUID)
}

type ClusterFeatures struct {
	NodeCount            int      `mapstructure:"node-count"`
	CellGUIDs            []string `mapstructure:"cells"`
	CloneFromServiceName string   `mapstructure:"clone-from"`
}

type PostgresCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Node struct {
	ID       string `json:"node_id"`
	CellGUID string `json:"cell_guid"`
	State    string `json:"state"`
	Role     string `json:"role"`
}

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
