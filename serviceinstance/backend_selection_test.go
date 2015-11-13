package serviceinstance_test

import (
	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/config"
	"github.com/cloudfoundry-community/patroni-broker/serviceinstance"
	"github.com/frodenas/brokerapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

const small = 1
const medium = 2

// SetFakeSize is used by request_tests to set an initial cluster size before changes
func setFakeSize(cluster *serviceinstance.Cluster, nodeCount, nodeSize int) {
	cluster.NodeCount = nodeCount
	cluster.NodeSize = nodeSize
}

var _ = Describe("backend broker selection", func() {
	var etcdClient backend.FakeEtcdClient
	cfg := &config.Config{}
	var logger lager.Logger
	clusterUUID := "uuid"
	var cluster *serviceinstance.Cluster
	var serviceDetails brokerapi.ProvisionDetails

	BeforeEach(func() {
		cfg.Backends = []*config.Backend{
			&config.Backend{AvailabilityZone: "z1"},
			&config.Backend{AvailabilityZone: "z1"},
			&config.Backend{AvailabilityZone: "z1"},
			&config.Backend{AvailabilityZone: "z2"},
			&config.Backend{AvailabilityZone: "z2"},
			&config.Backend{AvailabilityZone: "z2"},
		}
	})

	It("has 3 backends", func() {
		Ω(len(cfg.Backends)).To(Equal(6))
	})

	Context("cluster change from 0 to 1", func() {
		BeforeEach(func() {
			cluster = serviceinstance.NewCluster(clusterUUID, serviceDetails, etcdClient, cfg, logger)
			setFakeSize(cluster, 0, small)
		})
		It("is initial creation", func() {
			backends := cluster.SortedBackendsByUnusedAZs()
			Ω(len(backends)).To(Equal(6))
		})
	})
})
