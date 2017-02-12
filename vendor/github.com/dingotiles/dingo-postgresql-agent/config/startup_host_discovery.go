package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// HostDiscoverySpecification describes the host machine's ports mapped into this agent's container
type HostDiscoverySpecification struct {
	IP       string `required:"true" envconfig:"docker_host_ip"`
	Port5432 string `default:"5432" envconfig:"docker_host_port_5432"`
	Port8008 string `default:"8008" envconfig:"docker_host_port_8008"`
}

var hostDiscoverySpec *HostDiscoverySpecification

// HostDiscoverySpec discovers the host machine's ports mapped into this agent's container
// from env vars that prefix with DOCKER_HOST_
// E.g. DOCKER_HOST_IP, DOCKER_HOST_PORT_5432, DOCKER_HOST_PORT_8008
func HostDiscoverySpec() *HostDiscoverySpecification {
	if hostDiscoverySpec == nil {
		hostDiscoverySpec = &HostDiscoverySpecification{}
		err := envconfig.Process("docker_host", hostDiscoverySpec)
		if err != nil {
			log.Fatal(err.Error())
		}
		// TODO: Display warning if using default ports
		// TODO: Verify that DOCKER_HOST_IP:5432 connects to running postgresql
		// TODO: Verify that DOCKER_HOST_IP:8008 connects to running patroni api
	}
	return hostDiscoverySpec
}
