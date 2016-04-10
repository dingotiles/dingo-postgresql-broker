package serviceinstance

import "github.com/dingotiles/dingo-postgresql-broker/config"

// SortedBackendsByUnusedAZs is sequence of backends to try to request new nodes for this cluster
// It prioritizes backends in availability zones that are not currently used
func (cluster *Cluster) SortedBackendsByUnusedAZs() (backends []*config.Backend) {
	usedBackends, unusedBackeds := cluster.usedAndUnusedBackends()

	for _, az := range cluster.sortBackendAZsByUnusedness().Keys {
		for _, backend := range unusedBackeds {
			if backend.AvailabilityZone == az {
				backends = append(backends, backend)
			}
		}
	}
	for _, backend := range usedBackends {
		backends = append(backends, backend)
	}
	return
}

func (cluster *Cluster) usedAndUnusedBackends() (usedBackends, unusuedBackends []*config.Backend) {
	usedBackendGUIDs := cluster.usedBackendGUIDs()
	for _, backend := range cluster.AllBackends() {
		used := false
		for _, usedBackendGUID := range usedBackendGUIDs {
			if backend.GUID == usedBackendGUID {
				usedBackends = append(usedBackends, backend)
				used = true
				break
			}
		}
		if !used {
			unusuedBackends = append(unusuedBackends, backend)
		}
	}
	return
}
