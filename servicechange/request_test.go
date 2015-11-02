package servicechange_test

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const small = 1
const medium = 2

var _ = Describe("Service instance changes", func() {
	Describe(".Steps", func() {
		var req servicechange.RealRequest

		Context("no change", func() {
			BeforeEach(func() {
				cluster := serviceinstance.NewFakeCluster(1, small)
				req = servicechange.NewRequest(cluster)
			})
			It("noop", func() {
				Ω(req.Steps()).To(Equal([]servicechange.Step{}))
			})
		})

		Describe("grow cluster size (more replica nodes)", func() {
			Context("1 small master", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(1, small)
					req = servicechange.NewRequest(cluster)
				})
				It("adds (replica) node", func() {
					req.NewNodeCount = 2
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
				})
				It("adds multiple (replica) nodes", func() {
					req.NewNodeCount = 4
					steps := req.Steps()
					Ω(steps).To(HaveLen(3))
				})
			})
		})

		Describe("shrink cluster size (reduce replica nodes)", func() {
			Context("4-small nodes", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(4, small)
					req = servicechange.NewRequest(cluster)
				})
				It("remove (replica) node", func() {
					req.NewNodeCount = 3
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
				})
				It("removes (replica) nodes", func() {
					req.NewNodeCount = 1
					steps := req.Steps()
					Ω(steps).To(HaveLen(3))
				})
			})
		})

		Describe("resize cluster nodes (bigger or smaller nodes)", func() {
			Context("1-small becoming 1-medium", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(1, small)
					req = servicechange.NewRequest(cluster)
					req.NewNodeSize = medium
				})
				It("has steps", func() {
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
					Ω(steps[0]).To(BeAssignableToTypeOf(servicechange.StepReplaceMaster{}))
				})
			})
		})

		Describe("resize node and grow cluster count", func() {
			Context("2-small becoming 4-medium node cluster", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(2, small)
					req = servicechange.NewRequest(cluster)
					req.NewNodeCount = 4
					req.NewNodeSize = medium
				})
				It("has steps", func() {
					steps := req.Steps()
					Ω(steps).To(HaveLen(4))
					Ω(steps[0]).To(BeAssignableToTypeOf(servicechange.StepReplaceMaster{}))
					Ω(steps[1]).To(BeAssignableToTypeOf(servicechange.StepReplaceReplica{}))
					Ω(steps[2]).To(BeAssignableToTypeOf(servicechange.StepAddNode{}))
					Ω(steps[3]).To(BeAssignableToTypeOf(servicechange.StepAddNode{}))
				})
			})

			Context("6-medium becoming 3-small node cluster", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(6, medium)
					req = servicechange.NewRequest(cluster)
					req.NewNodeCount = 3
					req.NewNodeSize = small
				})
				It("has steps", func() {
					steps := req.Steps()
					Ω(steps).To(HaveLen(6))
					Ω(steps[0]).To(BeAssignableToTypeOf(servicechange.StepReplaceMaster{}))
					Ω(steps[1]).To(BeAssignableToTypeOf(servicechange.StepReplaceReplica{}))
					Ω(steps[2]).To(BeAssignableToTypeOf(servicechange.StepReplaceReplica{}))
					Ω(steps[3]).To(BeAssignableToTypeOf(servicechange.StepRemoveNode{}))
					Ω(steps[4]).To(BeAssignableToTypeOf(servicechange.StepRemoveNode{}))
					Ω(steps[5]).To(BeAssignableToTypeOf(servicechange.StepRemoveNode{}))
				})
			})

		})

	})
})
