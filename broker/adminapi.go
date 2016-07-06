package broker

import (
	"encoding/json"
	"net/http"

	"github.com/frodenas/brokerapi"
	"github.com/frodenas/brokerapi/auth"
	"github.com/pivotal-golang/lager"
)

func NewAdminAPI(serviceBroker *Broker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Get("/admin/cells", cells(serviceBroker, router, logger))
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

func cells(serviceBroker *Broker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		respond(w, http.StatusOK, serviceBroker.Cells())
	}
}
