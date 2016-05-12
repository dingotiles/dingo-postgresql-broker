package backend

import "github.com/dingotiles/dingo-postgresql-broker/config"

type Backend struct {
	Config *config.Backend
}

type Backends []*Backend

func NewBackend(config *config.Backend) *Backend {
	return &Backend{
		Config: config,
	}
}

// func (b *Backend) ProvisionNode(clusterData state.ClusterData) (nodeId string, err error) {
// 	nodeId = uuid.New()
// 	provisionDetails := brokerapi.ProvisionDetails{
// 		OrganizationGUID: clusterData.OrganizationGUID,
// 		PlanID:           clusterData.PlanID,
// 		ServiceID:        clusterData.Service D,
// 		SpaceGUID:        clusterData.SpaceGUID,
// 		Parameters: map[string]interface{}{
// 			"PATRONI_SCOPE":     clusterData.InstanceID,
// 			"NODE_NAME":         nodeId,
// 			"POSTGRES_USERNAME": clusterData.AdminCredentials.Username,
// 			"POSTGRES_PASSWORD": clusterData.AdminCredentials.Password,
// 		},
// 	}
// 	return
// }
