package scheduler_test

import (
	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/frodenas/brokerapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

const small = 1
const medium = 2

// SetFakeSize is used by request_tests to set an initial cluster size before changes
func setFakeSize(cluster *cluster.Cluster, nodeCount, nodeSize int) {
	cluster.Data.NodeCount = nodeCount
	cluster.Data.NodeSize = nodeSize
}

var _ = Describe("Service instance changes", func() {
	var etcdClient backend.FakeEtcdClient
	cfg := &bkrconfig.Config{}
	var logger lager.Logger

	Describe(".Steps", func() {
		var cluster *cluster.Cluster
		var req scheduler.Request
		var steps []step.Step
		clusterUUID := "uuid"
		var serviceDetails brokerapi.ProvisionDetails

		Context("no change", func() {
			BeforeEach(func() {
				cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
				setFakeSize(cluster, 1, small)
				req = scheduler.NewRequest(cluster, 1, small)
			})
			It("noop", func() {
				Ω(req.Steps()).To(Equal([]step.Step{}))
			})
		})

		Describe("new cluster", func() {
			Context("create 1 small master", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 0, small)
				})
				It("is initial creation", func() {
					req = scheduler.NewRequest(cluster, 1, small)
					Ω(req.IsInitialProvision()).To(BeTrue())
				})
				It("add master node", func() {
					req = scheduler.NewRequest(cluster, 1, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.AddNode{}))
				})
				It("adds master + replica", func() {
					req = scheduler.NewRequest(cluster, 2, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(2))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.AddNode{}))
					Ω(steps[1]).To(BeAssignableToTypeOf(step.AddNode{}))
				})
			})
		})
		Describe("destroy cluster", func() {
			Context("with 2-small", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 2, small)
				})
				It("removes all nodes", func() {
					req = scheduler.NewRequest(cluster, 0, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(2))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.RemoveNode{}))
					Ω(steps[1]).To(BeAssignableToTypeOf(step.RemoveNode{}))
				})
			})
		})

		Describe("grow cluster size (more replica nodes)", func() {
			Context("1 small master", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 1, small)
				})
				It("adds (replica) node", func() {
					req = scheduler.NewRequest(cluster, 2, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.AddNode{}))
				})
				It("adds multiple (replica) nodes", func() {
					req = scheduler.NewRequest(cluster, 4, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(3))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.AddNode{}))
					Ω(steps[1]).To(BeAssignableToTypeOf(step.AddNode{}))
					Ω(steps[2]).To(BeAssignableToTypeOf(step.AddNode{}))
				})
			})
		})

		Describe("shrink cluster size (reduce replica nodes)", func() {
			Context("4-small nodes", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 4, small)
				})
				It("remove (replica) node", func() {
					req = scheduler.NewRequest(cluster, 3, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
				})
				It("removes (replica) nodes", func() {
					req = scheduler.NewRequest(cluster, 1, small)
					steps := req.Steps()
					Ω(steps).To(HaveLen(3))
				})
			})
		})

		Describe("resize cluster nodes (bigger or smaller nodes)", func() {
			Context("1-small becoming 1-medium", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 1, small)
					req = scheduler.NewRequest(cluster, 1, medium)
				})
				It("has steps", func() {
					steps := req.Steps()
					Ω(steps).To(HaveLen(1))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.ReplaceMaster{}))
				})
			})
		})

		Describe("resize node and grow cluster count", func() {
			Context("2-small becoming 4-medium node cluster", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 2, small)
					req = scheduler.NewRequest(cluster, 4, medium)
				})
				It("has steps", func() {
					steps := req.Steps()
					Ω(steps).To(HaveLen(4))
					Ω(steps[0]).To(BeAssignableToTypeOf(step.ReplaceMaster{}))
					Ω(steps[1]).To(BeAssignableToTypeOf(step.ReplaceReplica{}))
					Ω(steps[2]).To(BeAssignableToTypeOf(step.AddNode{}))
					Ω(steps[3]).To(BeAssignableToTypeOf(step.AddNode{}))
				})
			})

			Context("6-medium becoming 3-small node cluster", func() {
				BeforeEach(func() {
					cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
					setFakeSize(cluster, 6, medium)
					req = scheduler.NewRequest(cluster, 3, small)
					steps = req.Steps()
				})
				It("is scaling down", func() {
					Ω(req.IsScalingDown()).To(BeTrue())
				})
				It("is scaling in", func() {
					Ω(req.IsScalingIn()).To(BeTrue())
				})
				It("has 6 steps", func() {
					Ω(steps).To(HaveLen(6))
				})
				It("step 0 replaces master", func() {
					Ω(steps[0]).To(BeAssignableToTypeOf(step.ReplaceMaster{}))
				})
				It("step 1 & 2 replaces replica", func() {
					Ω(steps[1]).To(BeAssignableToTypeOf(step.ReplaceReplica{}))
					Ω(steps[2]).To(BeAssignableToTypeOf(step.ReplaceReplica{}))
				})
				It("step 3,4,5 removes other replicas", func() {
					Ω(steps[3]).To(BeAssignableToTypeOf(step.RemoveNode{}))
					Ω(steps[4]).To(BeAssignableToTypeOf(step.RemoveNode{}))
					Ω(steps[5]).To(BeAssignableToTypeOf(step.RemoveNode{}))
				})
			})

		})

	})
})
