package structs

import (
	"testing"
)

func TestClusterData_Equals(t *testing.T) {
	data := &ClusterData{
		InstanceID:       "instanceID",
		OrganizationGUID: "OrganizationGUID",
		PlanID:           "PlanID",
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
		AdminCredentials: PostgresCredentials{
			Username: "pgadmin",
			Password: "pw",
		},
	}
	otherData := ClusterData{
		InstanceID:       "instanceID",
		OrganizationGUID: "OrganizationGUID",
		PlanID:           "PlanID",
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
		AdminCredentials: PostgresCredentials{
			Username: "pgadmin",
			Password: "pw",
		},
	}
	if otherData.Equals(data) != true {
		t.Errorf("ClusterData %v should be equal to %v", otherData, data)
	}
	if data.Equals(&otherData) != true {
		t.Errorf("ClusterData %v should be equal to %v", data, otherData)
	}

	otherData.InstanceID = "otherID"

	if otherData.Equals(data) == true {
		t.Errorf("ClusterData %v should not be equal to %v", otherData, data)
	}
	if data.Equals(&otherData) == true {
		t.Errorf("ClusterData %v should not be equal to %v", data, otherData)
	}
}
