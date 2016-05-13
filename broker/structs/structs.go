package structs

import "reflect"

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
	Id        string
	BackendId string
	PlanId    string
	ServiceId string
}

func (data *ClusterData) Equals(other *ClusterData) bool {
	return reflect.DeepEqual(*data, *other)
}
