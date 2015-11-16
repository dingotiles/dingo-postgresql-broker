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

var _ = Describe("backend broker selection", func() {
	etcdClient := backend.NewFakeEtcdClient()
	cfg := &config.Config{}
	logger := lager.NewLogger("tests")
	clusterUUID := "uuid"
	var cluster *serviceinstance.Cluster
	var serviceDetails brokerapi.ProvisionDetails

	It("has three AZs", func() {
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
		cluster = serviceinstance.NewCluster(clusterUUID, serviceDetails, etcdClient, cfg, logger)
		Ω(cluster.AllAZs()).To(Equal([]string{"z1", "z2", "z3"}))
	})
})
