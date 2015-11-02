package servicechange_test

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service instance changes", func() {
	Describe(".Steps", func() {
		var req servicechange.RealRequest

		Context("no change", func() {
			BeforeEach(func() {
				cluster := serviceinstance.NewFakeCluster(1, "small")
				req = servicechange.NewRequest(cluster)
			})
			It("noop", func() {
				Ω(req.Steps()).To(Equal([]servicechange.Step{}))
			})
		})

		Describe("grow cluster size (more replica nodes)", func() {
			Context("1 small master", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(1, "small")
					req = servicechange.NewRequest(cluster)
				})
				It("adds (replica) node", func() {
					req.ChangeNodeCount = +1
					steps := req.Steps()
					Ω(len(steps)).To(Equal(1))
				})
				It("adds multiple (replica) nodes", func() {
					req.ChangeNodeCount = +3
					steps := req.Steps()
					Ω(len(steps)).To(Equal(3))
				})
			})
		})

		Describe("shrink cluster size (reduce replica nodes)", func() {
			Context("4-small nodes", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(4, "small")
					req = servicechange.NewRequest(cluster)
				})
				It("remove (replica) node", func() {
					req.ChangeNodeCount = -1
					steps := req.Steps()
					Ω(len(steps)).To(Equal(1))
				})
				It("removes (replica) nodes", func() {
					req.ChangeNodeCount = -3
					steps := req.Steps()
					Ω(len(steps)).To(Equal(3))
				})
			})
		})

		XDescribe("resize cluster nodes (bigger or smaller nodes)", func() {
			Context("1-small node cluster", func() {
				BeforeEach(func() {
					cluster := serviceinstance.NewFakeCluster(1, "small")
					req = servicechange.NewRequest(cluster)
				})
				Context("single node cluster", func() {
					It("has steps", func() {
						req.ChangeNodeSize = "mega-large"
						steps := req.Steps()
						Ω(len(steps)).To(Equal(2))
						// 1. add replica
						// 2. kill master; force promote a replica
					})
				})

			})
		})

	})
})
