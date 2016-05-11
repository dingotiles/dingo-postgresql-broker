package cluster_test

import (
	"fmt"

	"github.com/coreos/go-etcd/etcd"
	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/cluster"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/frodenas/brokerapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

const small = 1
const medium = 2

// SetFakeSize is used by request_tests to set an initial cluster size before changes
func setupCluster(cluster *cluster.Cluster, existingBackendGUIDs []string) {
	etcdClient := cluster.etcdClient.(backend.FakeEtcdClient)

	nodes := make(etcd.Nodes, len(existingBackendGUIDs))
	for i, backendGUID := range existingBackendGUIDs {
		key := fmt.Sprintf("/serviceinstances/%s/nodes/instance-%d/backend", cluster.Data.InstanceID, i)
		etcdClient.GetResponses[key] = &etcd.Response{
			Node: &etcd.Node{
				Key:   key,
				Value: backendGUID,
			},
		}
		nodes[i] = &etcd.Node{Key: fmt.Sprintf("/serviceinstances/%s/nodes/instance-%d", cluster.Data.InstanceID, i)}
	}

	key := fmt.Sprintf("/serviceinstances/%s/nodes", cluster.Data.InstanceID)
	etcdClient.GetResponses[key] = &etcd.Response{
		Node: &etcd.Node{Key: key, Nodes: nodes},
	}

}

func backendGUIDs(backends []*config.Backend) []string {
	guids := make([]string, len(backends))
	for i, backend := range backends {
		guids[i] = backend.GUID
	}
	return guids
}

var _ = Describe("backend broker selection", func() {
	etcdClient := backend.NewFakeEtcdClient()
	cfg := &config.Config{}
	var logger lager.Logger
	clusterUUID := "uuid"
	var cluster *cluster.Cluster
	var serviceDetails brokerapi.ProvisionDetails
	var backends []*config.Backend

	Context("two AZs", func() {
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

		It("orders backends before cluster change from 0 to 1", func() {
			cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
			setupCluster(cluster, []string{})
			backends := cluster.SortedBackendsByUnusedAZs()
			Ω(backendGUIDs(backends)).To(Equal([]string{"c1z1", "c2z1", "c3z1", "c4z2", "c5z2", "c6z2"}))
		})

		Context("orders backends before cluster change from 1 to 2", func() {
			It("has list of backends with z2 first, z1 second, c1z1 last", func() {
				cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
				setupCluster(cluster, []string{"c1z1"})
				backends = cluster.SortedBackendsByUnusedAZs()
				// backend broker already used is last in the list; its AZ is the last AZ
				// example expected result (z2's first, then z1's):
				//   "c4z2", "c5z2", "c6z2", "c2z1", "c3z1", "c1z1"
				Ω(backends[0].GUID).To(MatchRegexp("z2$"))
				Ω(backends[1].GUID).To(MatchRegexp("z2$"))
				Ω(backends[2].GUID).To(MatchRegexp("z2$"))
				Ω(backends[3].GUID).To(MatchRegexp("z1$"))
				Ω(backends[4].GUID).To(MatchRegexp("z1$"))
				Ω(backends[5].GUID).To(Equal("c1z1"))
			})
		})

		Context("orders backends before cluster change from 2 to 3", func() {
			It("has list of backends c1z1,c4z2 last", func() {
				cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
				setupCluster(cluster, []string{"c1z1", "c4z2"})
				backends = cluster.SortedBackendsByUnusedAZs()
				// backend broker already used is last in the list; its AZ is the last AZ
				Ω(backendGUIDs(backends)).To(Equal([]string{"c2z1", "c3z1", "c5z2", "c6z2", "c1z1", "c4z2"}))
			})
		})
	})

	Context("three AZs", func() {
		BeforeEach(func() {
			cfg.Backends = []*config.Backend{
				&config.Backend{AvailabilityZone: "z1", GUID: "c1z1"},
				&config.Backend{AvailabilityZone: "z1", GUID: "c2z1"},
				&config.Backend{AvailabilityZone: "z1", GUID: "c3z1"},
				&config.Backend{AvailabilityZone: "z2", GUID: "c4z2"},
				&config.Backend{AvailabilityZone: "z2", GUID: "c5z2"},
				&config.Backend{AvailabilityZone: "z2", GUID: "c6z2"},
				&config.Backend{AvailabilityZone: "z3", GUID: "c7z3"},
				&config.Backend{AvailabilityZone: "z3", GUID: "c8z3"},
				&config.Backend{AvailabilityZone: "z3", GUID: "c9z3"},
			}
		})

		It("orders backends before cluster change from 0 to 1", func() {
			cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
			setupCluster(cluster, []string{})
			backends := cluster.SortedBackendsByUnusedAZs()
			Ω(backendGUIDs(backends)).To(Equal([]string{"c1z1", "c2z1", "c3z1", "c4z2", "c5z2", "c6z2", "c7z3", "c8z3", "c9z3"}))
		})

		It("orders backends with 2 x z1 and 1 x z3 already in use", func() {
			cluster = cluster.NewClusterFromProvisionDetails(clusterUUID, serviceDetails, etcdClient, cfg, logger)
			setupCluster(cluster, []string{"c1z1", "c2z1", "c7z3"})
			backends = cluster.SortedBackendsByUnusedAZs()
			// backend broker already used is last in the list; its AZ is the last AZ
			fmt.Printf("%#v\n", backendGUIDs(backends))
			Ω(backends[0].GUID).To(MatchRegexp("z2$")) // no z2 used so far; so should be first in list
			Ω(backends[1].GUID).To(MatchRegexp("z2$"))
			Ω(backends[2].GUID).To(MatchRegexp("z2$"))
			Ω(backends[3].GUID).To(MatchRegexp("z3$")) // z3 is middle
			Ω(backends[4].GUID).To(MatchRegexp("z3$"))
			Ω(backends[5].GUID).To(Equal("c3z1")) // the last z1; and z1 is most overused AZ
			Ω(backends[6].GUID).To(Equal("c1z1"))
			Ω(backends[7].GUID).To(Equal("c2z1"))
			Ω(backends[8].GUID).To(Equal("c7z3"))
		})
	})

})
