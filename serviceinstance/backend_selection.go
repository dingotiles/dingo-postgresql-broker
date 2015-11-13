package serviceinstance

import "github.com/cloudfoundry-community/patroni-broker/config"

// SortedBackendsByUnusedAZs is sequence of backends to try to request new nodes for this cluster
// It prioritizes backends in availability zones that are not currently used
func (cluster *Cluster) SortedBackendsByUnusedAZs() (backends []*config.Backend) {
	for _, az := range cluster.sortBackendAZsByUnusedness().Keys {
		for _, backend := range cluster.AllBackends() {
			if backend.AvailabilityZone == az {
				backends = append(backends, backend)
			}
		}
	}
	return
}
