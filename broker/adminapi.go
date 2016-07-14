package broker

import (
	"encoding/json"
	"net/http"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/frodenas/brokerapi/auth"
	"github.com/pivotal-golang/lager"
)

func NewAdminAPI(serviceBroker *Broker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Get("/admin/cells", adminCells(serviceBroker, router, logger))
	router.Get("/admin/service_instances/{instance_id}", adminServiceInstances(serviceBroker, router, logger))
	return wrapAuth(router, brokerCredentials)
}

func wrapAuth(router httpRouter, credentials brokerapi.BrokerCredentials) http.Handler {
	return auth.NewWrapper(credentials.Username, credentials.Password).Wrap(router)
}

func respond(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

type adminCell struct {
	GUID             string `json:"guid"`
	AvailabilityZone string `json:"availability_zone"`
	URI              string `json:"uri"`
	Username         string `json:"username"`
	Password         string `json:"password"`
}

func adminCells(bkr *Broker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := bkr.newLoggingSession("admin.cells", lager.Data{})
		defer logger.Info("done")

		cells := bkr.Cells()
		resultCells := []*adminCell{}

		for _, cell := range cells {
			adminCell := adminCell{
				GUID:             cell.GUID,
				AvailabilityZone: cell.AvailabilityZone,
				URI:              cell.URI,
				Username:         cell.Username,
				Password:         cell.Password,
			}
			resultCells = append(resultCells, &adminCell)
		}

		respond(w, http.StatusOK, resultCells)
	}
}

func adminServiceInstances(bkr *Broker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := structs.ClusterID(vars["instance_id"])

		logger := bkr.newLoggingSession("admin.service-instances", lager.Data{"instance-id": instanceID})
		defer logger.Info("done")

		cluster, err := bkr.state.LoadCluster(instanceID)
		if err != nil {
			logger.Error("load-cluster.error", err)
			respond(w, http.StatusInternalServerError, err.Error())
		}

		respond(w, http.StatusOK, cluster)
	}
}
