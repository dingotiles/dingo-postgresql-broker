package adminapi

import (
	"fmt"
	"net/http"

	"github.com/frodenas/brokerapi"
	"github.com/frodenas/brokerapi/auth"
	"github.com/pivotal-golang/lager"
)

func New(serviceBroker brokerapi.ServiceBroker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Get("/admin/hello/", hello(serviceBroker, router, logger))
	return wrapAuth(router, brokerCredentials)
}

func wrapAuth(router httpRouter, credentials brokerapi.BrokerCredentials) http.Handler {
	return auth.NewWrapper(credentials.Username, credentials.Password).Wrap(router)
}

func hello(serviceBroker brokerapi.ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		message := vars["message"]

		w.Write([]byte(fmt.Sprintf("admin %s", message)))
	}
}
