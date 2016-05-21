package structs

import (
	"encoding/json"
	"reflect"
)

type ClusterRecreationData struct {
	InstanceID       string           `json:"instance_id"`
	ServiceID        string           `json:"service_id"`
	PlanID           string           `json:"plan_id"`
	OrganizationGUID string           `json:"organization_guid"`
	SpaceGUID        string           `json:"space_guid"`
	AdminCredentials AdminCredentials `json:"admin_credentials"`
	AllocatedPort    int              `json:"allocated_port"`
}

type ClusterState struct {
	InstanceID       string           `json:"instance_id"`
	ServiceID        string           `json:"service_id"`
	PlanID           string           `json:"plan_id"`
	OrganizationGUID string           `json:"organization_guid"`
	SpaceGUID        string           `json:"space_guid"`
	AdminCredentials AdminCredentials `json:"admin_credentials"`
	AllocatedPort    int              `json:"allocated_port"`
	Nodes            []*Node          `json:"nodes"`
}

func (c *ClusterState) NodeCount() int {
	return len(c.Nodes)
}

func (c *ClusterState) RecreationData() *ClusterRecreationData {
	return &ClusterRecreationData{
		InstanceID:       c.InstanceID,
		ServiceID:        c.ServiceID,
		PlanID:           c.PlanID,
		OrganizationGUID: c.OrganizationGUID,
		SpaceGUID:        c.SpaceGUID,
		AdminCredentials: c.AdminCredentials,
		AllocatedPort:    c.AllocatedPort,
	}
}

type Cluster interface {
	AllNodes() []*Node
	AddNode(Node) error
	RemoveNode(*Node) error
	MetaData() ClusterData
}

func (c *ClusterState) AllNodes() []*Node {
	return c.Nodes
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

func (c *ClusterState) MetaData() ClusterData {
	return ClusterData{
		InstanceID:       c.InstanceID,
		ServiceID:        c.ServiceID,
		PlanID:           c.PlanID,
		OrganizationGUID: c.OrganizationGUID,
		SpaceGUID:        c.SpaceGUID,
		AdminCredentials: c.AdminCredentials,
		AllocatedPort:    c.AllocatedPort,
	}
}

type ClusterFeatures struct {
	NodeCount int `json:"node_count"`
}

type AdminCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ClusterData describes the current request for the state of the cluster
type ClusterData struct {
	InstanceID       string           `json:"instance_id"`
	ServiceID        string           `json:"service_id"`
	PlanID           string           `json:"plan_id"`
	OrganizationGUID string           `json:"organization_guid"`
	SpaceGUID        string           `json:"space_guid"`
	AdminCredentials AdminCredentials `json:"admin_credentials"`
	TargetNodeCount  int              `json:"node_count"`
	AllocatedPort    int              `json:"allocated_port"`
}

type Node struct {
	ID        string `json:"node_id"`
	BackendID string `json:"backend_id"`
	PlanID    string `json:"plan_id"`
	ServiceID string `json:"service_id"`
}

func (data *ClusterData) Equals(other *ClusterData) bool {
	return reflect.DeepEqual(*data, *other)
}

func (c *ClusterData) Json() string {
	data, _ := json.Marshal(c)
	return (string(data))
}

func ClusterDataFromJson(jsonString string) *ClusterData {
	data := ClusterData{}
	json.Unmarshal([]byte(jsonString), &data)
	return &data
}
