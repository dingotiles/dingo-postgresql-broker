package broker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/frodenas/brokerapi/auth"
	"github.com/pivotal-golang/lager"
)

func NewAdminAPI(serviceBroker *Broker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Post("/admin/cells/{cell_guid}/demote", demoteCell(serviceBroker, router, logger))
	router.Get("/admin/cells", adminCells(serviceBroker, router, logger))
	router.Get("/admin/service_instances/{instance_id}", adminServiceInstances(serviceBroker, router, logger))
	router.Get("/admin/spaces/{space_guid}/clusterdata_backup_by_name/{name}", adminFindServiceInstanceByName(serviceBroker, router, logger))
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

func demoteCell(bkr *Broker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		cellGUID := vars["cell_guid"]

		logger := bkr.newLoggingSession("admin.cells.demote", lager.Data{"cell-guid": cellGUID})
		defer logger.Info("done")

		allClusters, err := bkr.state.LoadAllRunningClusters()
		if err != nil {
			logger.Error("load-clusters.error", err)
			respond(w, http.StatusInternalServerError, err.Error())
			return
		}

		var wg sync.WaitGroup
		for _, cluster := range allClusters {
			wg.Add(1)
			thisCluster := cluster
			go func() {
				defer wg.Done()

				node, err := thisCluster.NodeOnCell(cellGUID)
				if err != nil {
					logger.Error("node-on-cell.error", err)
					return
				}

				err = bkr.patroni.FailoverFrom(thisCluster.InstanceID, node.ID)
				if err != nil {
					logger.Error("failover.error",
						fmt.Errorf("Couldn't failover member %s from instance %s: '%s'",
							node.ID, thisCluster.InstanceID, err))
					return
				}
				logger.Info("failover.success", lager.Data{"instance-id": thisCluster.InstanceID, "member-id": node.ID})
			}()
		}
		wg.Wait()
		respond(w, http.StatusOK, fmt.Sprintf("Failover from cell %s completed", cellGUID))
	}
}

func adminFindServiceInstanceByName(bkr *Broker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		spaceGUID := string(vars["space_guid"])
		name := string(vars["name"])

		logger := bkr.newLoggingSession("admin.find-service-instance-by-name",
			lager.Data{"space-guid": spaceGUID, "name": name})
		defer logger.Info("done")

		data, err := bkr.callbacks.ClusterDataFindServiceInstanceByName(spaceGUID, name)
		if err != nil {
			logger.Error("error", err)
			respond(w, http.StatusInternalServerError, err.Error())
			return
		}

		respond(w, http.StatusNotFound, data)
	}
}
