package structs

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	defaultNodeCount = 2
)

type ClusterID string

type ClusterRecreationData struct {
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
	Nodes                []*Node             `json:"nodes"`
	ErrorMsg             string              `json:"error"`
}

func (c *ClusterState) NodeCount() int {
	return len(c.Nodes)
}

func (c *ClusterState) RecreationData() *ClusterRecreationData {
	return &ClusterRecreationData{
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

func (c *ClusterState) AddNode(node Node) error {
	c.Nodes = append(c.Nodes, &node)
	return nil
}

func (c *ClusterState) RemoveNode(node *Node) error {
	for i, n := range c.Nodes {
		if n.ID == node.ID {
			c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
			break
		}
	}
	return nil
}

type ClusterFeatures struct {
	NodeCount int      `mapstructure:"node-count"`
	CellGUIDs []string `mapstructure:"cell-guids"`
}

type PostgresCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Node struct {
	ID        string `json:"node_id"`
	BackendID string `json:"backend_id"`
	PlanID    string `json:"plan_id"`
	ServiceID string `json:"service_id"`
	Role      string `json:"role"`
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
