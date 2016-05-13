package structs

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
		}
		Ω(otherData.Equals(data)).To(Equal(true))
		Ω(data.Equals(&otherData)).To(Equal(true))

		otherData.InstanceID = "otherID"

		Ω(otherData.Equals(data)).To(Equal(false))
		Ω(data.Equals(&otherData)).To(Equal(false))
	})
})
