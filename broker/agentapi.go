package broker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	agentconfig "github.com/dingotiles/dingo-postgresql-agent/config"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// NewAgentAPI setups all API interfaces from dingo-postgresql-agent containers
func NewAgentAPI(serviceBroker *Broker, logger lager.Logger, brokerCredentials brokerapi.BrokerCredentials) http.Handler {
	router := newHTTPRouter()

	router.Post("/agent/api", agentStartRequest(serviceBroker, router, logger))
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
		clusterSpec.Cluster.Name = startupReq.NodeName
		clusterSpec.Cluster.Scope = startupReq.ClusterName

		arch := &clusterSpec.Archives
		archConfig := bkr.global.Archives
		arch.Method = archConfig.Method
		arch.S3.AWSAccessKeyID = archConfig.S3.AWSAccessKeyID
		arch.S3.AWSSecretAccessID = archConfig.S3.AWSSecretAccessID
		arch.S3.S3Bucket = archConfig.S3.S3Bucket
		arch.S3.S3Endpoint = awsRegionToS3Endpoint(archConfig.S3.AWSRegion)
		arch.SSH.Host = archConfig.SSH.Host
		arch.SSH.Port = archConfig.SSH.Port
		arch.SSH.User = archConfig.SSH.User
		arch.SSH.BasePath = archConfig.SSH.BasePath
		arch.SSH.PrivateKey = archConfig.SSH.PrivateKey

		clusterSpec.Etcd.URI = bkr.global.Etcd.URI()

		instanceID := structs.ClusterID(startupReq.ClusterName)
		clusterState, err := bkr.state.LoadCluster(instanceID)
		if err != nil {
			respond(w, http.StatusBadRequest, fmt.Sprintf("Could not load cluster details: %s", err))
			return
		}

		clusterSpec.Postgresql.Admin.Password = clusterState.AdminCredentials.Password
		clusterSpec.Postgresql.Superuser.Username = clusterState.SuperuserCredentials.Username
		clusterSpec.Postgresql.Superuser.Password = clusterState.SuperuserCredentials.Password
		clusterSpec.Postgresql.Appuser.Username = clusterState.AppCredentials.Username
		clusterSpec.Postgresql.Appuser.Password = clusterState.AppCredentials.Password

		respond(w, http.StatusOK, clusterSpec)
	}
}

func requiredEnv(envKey string) string {
	if os.Getenv(envKey) == "" {
		missingRequiredEnvs = append(missingRequiredEnvs, envKey)
	}
	return os.Getenv(envKey)
}

// http://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
func awsRegionToS3Endpoint(awsRegion string) string {
	if awsRegion == "" {
		return ""
	}
	switch awsRegion {
	case "us-east-1":
		return "https+path://s3.amazonaws.com:443"
	case "ap-northeast-1":
		return "https+path://s3-ap-northeast-1.amazonaws.com:443"
	case "sa-east-1":
		return "https+path://s3-sa-east-1.amazonaws.com:443"
	}
	return fmt.Sprintf("https+path://s3.%s.amazonaws.com:443", awsRegion)
}
