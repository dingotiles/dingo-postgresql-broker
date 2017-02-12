package cells

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

type Cell struct {
	GUID             string
	URI              string
	Config           *config.Cell
	AvailabilityZone string
	clusterLoader    ClusterLoader
}

type Cells []*Cell

func NewCells(configs []*config.Cell, clusterLoader ClusterLoader) Cells {
	var cells []*Cell
	for _, cfg := range configs {
		cells = append(cells, newCell(cfg, clusterLoader))
	}
	return cells
}

func newCell(config *config.Cell, clusterLoader ClusterLoader) *Cell {
	return &Cell{
		GUID:             config.GUID,
		Config:           config,
		AvailabilityZone: config.AvailabilityZone,
		URI:              config.URI,
		clusterLoader:    clusterLoader,
	}
}

func (cells Cells) String() string {
	ids := make([]string, len(cells))
	for i, cell := range cells {
		ids[i] = cell.GUID
	}
	return fmt.Sprintf("%v", ids)
}

func (cells Cells) AllAvailabilityZones() []string {
	azMap := map[string]string{}
	for _, cell := range cells {
		azMap[cell.AvailabilityZone] = cell.AvailabilityZone
	}

	keys := make([]string, 0, len(azMap))
	for k := range azMap {
		keys = append(keys, k)
	}
	return keys
}

func (cells Cells) AvailabilityZone(cellID string) (string, error) {
	for _, cell := range cells {
		if cell.GUID == cellID {
			return cell.AvailabilityZone, nil
		}
	}
	return "", errors.New(fmt.Sprintf("No cell with ID %s found", cellID))
}

func (cells Cells) Get(cellID string) *Cell {
	for _, cell := range cells {
		if cell.GUID == cellID {
			return cell
		}
	}
	return nil
}

func (cells Cells) ContainsCell(cellID string) bool {
	for _, cell := range cells {
		if cell.GUID == cellID {
			return true
		}
	}
	return false
}

func (cell *Cell) ProvisionNode(clusterState structs.ClusterState, logger lager.Logger) (node structs.Node, err error) {
	node = structs.Node{ID: uuid.New(), CellGUID: cell.GUID}
	provisionDetails := brokerapi.ProvisionDetails{
		OrganizationGUID: clusterState.OrganizationGUID,
		PlanID:           clusterState.PlanID,
		ServiceID:        clusterState.ServiceID,
		SpaceGUID:        clusterState.SpaceGUID,
		Parameters: map[string]interface{}{
			"PATRONI_SCOPE":   clusterState.InstanceID,
			"NODE_ID":         node.ID,
			"DINGO_CLUSTER":   clusterState.InstanceID,
			"DINGO_ORG_TOKEN": "required-but-not-used",
		},
	}

	url := fmt.Sprintf("%s/v2/service_instances/%s", cell.Config.URI, node.ID)
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	if err = json.NewEncoder(buffer).Encode(provisionDetails); err != nil {
		logger.Error("request-node.cell-provision-encode-details", err)
		return
	}
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		logger.Error("request-node.cell-provision-req", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cell.Config.Username, cell.Config.Password)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("request-node.cell-provision-resp", err)
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

func (cell *Cell) DeprovisionNode(clusterState structs.ClusterState, node *structs.Node, logger lager.Logger) (err error) {
	url := fmt.Sprintf("%s/v2/service_instances/%s", cell.URI, node.ID)
	client := &http.Client{}
	buffer := &bytes.Buffer{}

	deleteDetails := brokerapi.DeprovisionDetails{
		PlanID:    clusterState.PlanID,
		ServiceID: clusterState.ServiceID,
	}

	if err = json.NewEncoder(buffer).Encode(deleteDetails); err != nil {
		logger.Error("remove-node.cell.encode", err)
		return err
	}
	req, err := http.NewRequest("DELETE", url, buffer)
	if err != nil {
		logger.Error("remove-node.cell.new-req", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cell.Config.Username, cell.Config.Password)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("remove-node.cell.do", err)
		return err
	}
	defer resp.Body.Close()

	return
}
