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

func backendGUIDs(backends []*config.Backend) []string {
	guids := make([]string, len(backends))
	for i, backend := range backends {
		guids[i] = backend.GUID
	}
	return guids
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
			&config.Backend{AvailabilityZone: "z1", GUID: "c1z1"},
			&config.Backend{AvailabilityZone: "z1", GUID: "c2z1"},
			&config.Backend{AvailabilityZone: "z1", GUID: "c3z1"},
			&config.Backend{AvailabilityZone: "z2", GUID: "c4z2"},
			&config.Backend{AvailabilityZone: "z2", GUID: "c5z2"},
			&config.Backend{AvailabilityZone: "z2", GUID: "c6z2"},
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
			Ω(backendGUIDs(backends)).To(Equal([]string{"c1z1", "c2z1", "c3z1", "c4z2", "c5z2", "c6z2"}))
		})
	})
})
