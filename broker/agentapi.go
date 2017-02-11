package broker

import (
	"encoding/json"
	"net/http"
	"os"

	agentconfig "github.com/dingotiles/dingo-postgresql-agent/config"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// NewAgentAPI setups all API interfaces from dingo-postgresql-agent containers
func NewAgentAPI(serviceBroker *Broker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Post("/agent/", agentStartRequest(serviceBroker, router, logger))
	return wrapAuth(router, brokerCredentials)
}

var missingRequiredEnvs = []string{}

func agentStartRequest(bkr *Broker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := bkr.newLoggingSession("agent.start", lager.Data{})
		defer logger.Info("done")

		startupReq := &agentconfig.ContainerStartupRequest{}
		// TODO marshal req into startupReq
		err := json.NewDecoder(req.Body).Decode(startupReq)
		if err != nil {
			respond(w, http.StatusBadRequest, "Missing startup request parameters")
			return
		}

		clusterSpec := agentconfig.ClusterSpecification{}
		// TODO - need UUID for containers within cluster
		// Name vs Scope?
		clusterSpec.Cluster.Name = startupReq.ClusterName
		clusterSpec.Cluster.Scope = startupReq.ClusterName

		arch := &clusterSpec.Archives
		archConfig := bkr.global.Archives
		arch.Method = archConfig.Method
		arch.S3.AWSAccessKeyID = archConfig.S3.AWSAccessKeyID
		arch.S3.AWSSecretAccessID = archConfig.S3.AWSSecretAccessID
		arch.S3.S3Bucket = archConfig.S3.S3Bucket
		arch.S3.S3Endpoint = archConfig.S3.S3Endpoint
		arch.SSH.Host = archConfig.SSH.Host
		arch.SSH.Port = archConfig.SSH.Port
		arch.SSH.User = archConfig.SSH.User
		arch.SSH.BasePath = archConfig.SSH.BasePath
		arch.SSH.PrivateKey = archConfig.SSH.PrivateKey

		clusterSpec.Etcd.URI = bkr.global.Etcd.URI()

		// TODO: look these up; created by provision_endpoint initCluster
		clusterSpec.Postgresql.Admin.Password = "admin-password"
		clusterSpec.Postgresql.Superuser.Username = "superuser-username"
		clusterSpec.Postgresql.Superuser.Password = "superuser-password"
		clusterSpec.Postgresql.Appuser.Username = "appuser-username"
		clusterSpec.Postgresql.Appuser.Password = "appuser-password"

		respond(w, http.StatusOK, clusterSpec)
	}
}

func requiredEnv(envKey string) string {
	if os.Getenv(envKey) == "" {
		missingRequiredEnvs = append(missingRequiredEnvs, envKey)
	}
	return os.Getenv(envKey)
}
