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
	AllocatedPort    string           `json:"allocated_port"`
	nodes            []*Node          `json:"nodes"`
}

func (c *ClusterState) NodeCount() int {
	return len(c.nodes)
}

type Cluster interface {
	Nodes() []*Node
	AddNode(Node) error
	RemoveNode(*Node) error
	MetaData() ClusterData
}

func (c *ClusterState) Nodes() []*Node {
	return c.nodes
}

func (c *ClusterState) AddNode(node Node) error {
	c.nodes = append(c.nodes, &node)
	return nil
}

func (c *ClusterState) RemoveNode(node *Node) error {
	for i, n := range c.nodes {
		if n.ID == node.ID {
			c.nodes = append(c.nodes[:i], c.nodes[i+1:]...)
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
	AllocatedPort    string           `json:"allocated_port"`
}

type Node struct {
	ID        string
	BackendID string
	PlanID    string
	ServiceID string
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
