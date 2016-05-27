package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/frodenas/brokerapi"
	"github.com/pborman/uuid"
	"github.com/pivotal-golang/lager"
)

type Backend struct {
	ID               string
	URI              string
	Config           *config.Backend
	AvailabilityZone string
}

type Backends []*Backend

func NewBackend(config *config.Backend) *Backend {
	return &Backend{
		ID:               config.GUID,
		Config:           config,
		AvailabilityZone: config.AvailabilityZone,
		URI:              config.URI,
	}
}

func (b Backends) AllAvailabilityZones() []string {
	azMap := map[string]string{}
	for _, backend := range b {
		azMap[backend.Config.AvailabilityZone] = backend.Config.AvailabilityZone
	}

	keys := make([]string, 0, len(azMap))
	for k := range azMap {
		keys = append(keys, k)
	}
	return keys
}

func (b Backends) AvailabilityZone(backendID string) (string, error) {
	for _, backend := range b {
		if backend.ID == backendID {
			return backend.AvailabilityZone, nil
		}
	}
	return "", errors.New(fmt.Sprintf("No backend with ID %s found", backendID))
}

func (b Backends) Get(backendID string) *Backend {
	for _, backend := range b {
		if backend.ID == backendID {
			return backend
		}
	}
	return nil
}

func (b *Backend) ProvisionNode(clusterState *structs.ClusterState, logger lager.Logger) (node structs.Node, err error) {
	node = structs.Node{ID: uuid.New(), BackendID: b.ID, PlanID: clusterState.PlanID, ServiceID: clusterState.ServiceID}
	provisionDetails := brokerapi.ProvisionDetails{
		OrganizationGUID: clusterState.OrganizationGUID,
		PlanID:           clusterState.PlanID,
		ServiceID:        clusterState.ServiceID,
		SpaceGUID:        clusterState.SpaceGUID,
		Parameters: map[string]interface{}{
			"PATRONI_SCOPE":      clusterState.InstanceID,
			"NODE_NAME":          node.ID,
			"ADMIN_USERNAME":     clusterState.AdminCredentials.Username,
			"ADMIN_PASSWORD":     clusterState.AdminCredentials.Password,
			"SUPERUSER_USERNAME": clusterState.SuperuserCredentials.Username,
			"SUPERUSER_PASSWORD": clusterState.SuperuserCredentials.Password,
			"APPUSER_USERNAME":   clusterState.AppCredentials.Username,
			"APPUSER_PASSWORD":   clusterState.AppCredentials.Password,
		},
	}

	url := fmt.Sprintf("%s/v2/service_instances/%s", b.Config.URI, node.ID)
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	if err = json.NewEncoder(buffer).Encode(provisionDetails); err != nil {
		logger.Error("request-node.backend-provision-encode-details", err)
		return
	}
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		logger.Error("request-node.backend-provision-req", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(b.Config.Username, b.Config.Password)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("request-node.backend-provision-resp", err)
		return
	}
	defer resp.Body.Close()

	// FIXME: If resp.StatusCode not 200 or 201, then try next
	if resp.StatusCode >= 400 {
		// FIXME: allow return of this error to end user
		return structs.Node{}, errors.New("unknown plan")
	}
	return
}

func (b *Backend) DeprovisionNode(node *structs.Node, logger lager.Logger) (err error) {
	url := fmt.Sprintf("%s/v2/service_instances/%s", b.URI, node.ID)
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	deleteDetails := brokerapi.DeprovisionDetails{
		PlanID:    node.PlanID,
		ServiceID: node.ServiceID,
	}

	if err = json.NewEncoder(buffer).Encode(deleteDetails); err != nil {
		logger.Error("remove-node.backend.encode", err)
		return err
	}
	req, err := http.NewRequest("DELETE", url, buffer)
	if err != nil {
		logger.Error("remove-node.backend.new-req", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(b.Config.Username, b.Config.Password)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("remove-node.backend.do", err)
		return err
	}
	defer resp.Body.Close()

	return
}
