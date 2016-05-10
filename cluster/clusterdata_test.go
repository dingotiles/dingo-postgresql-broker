package cluster

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterData", func() {
	It("Can can be Equal", func() {
		data := &ClusterData{
			InstanceID:       "instanceID",
			OrganizationGUID: "OrganizationGUID",
			PlanID:           "PlanID",
			ServiceID:        "ServiceID",
			SpaceGUID:        "SpaceGUID",
			AdminCredentials: AdminCredentials{
				Username: "pgadmin",
				Password: "pw",
			},
			Parameters: nil,
		}
		otherData := ClusterData{
			InstanceID:       "instanceID",
			OrganizationGUID: "OrganizationGUID",
			PlanID:           "PlanID",
			ServiceID:        "ServiceID",
			SpaceGUID:        "SpaceGUID",
			AdminCredentials: AdminCredentials{
				Username: "pgadmin",
				Password: "pw",
			},
			Parameters: nil,
		}
		Ω(otherData.Equals(data)).To(Equal(true))
		Ω(data.Equals(&otherData)).To(Equal(true))

		otherData.InstanceID = "otherID"

		Ω(otherData.Equals(data)).To(Equal(false))
		Ω(data.Equals(&otherData)).To(Equal(false))
	})
})
