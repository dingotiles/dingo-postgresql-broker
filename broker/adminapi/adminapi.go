package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/frodenas/brokerapi"
	"github.com/frodenas/brokerapi/auth"
	"github.com/pivotal-golang/lager"
)

func New(serviceBroker brokerapi.ServiceBroker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Get("/admin/hello/{person}", hello(serviceBroker, router, logger))
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

func hello(serviceBroker brokerapi.ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		person := vars["person"]

		respond(w, http.StatusOK, fmt.Sprintf("hello, %s", person))
	}
}
