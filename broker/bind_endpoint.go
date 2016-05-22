package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// CredentialsHash represents the set of binding credentials returned
type CredentialsHash struct {
	Host              string `json:"host,omitempty"`
	Port              int    `json:"port,omitempty"`
	Name              string `json:"name,omitempty"`
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	URI               string `json:"uri,omitempty"`
	JDBCURI           string `json:"jdbcUrl,omitempty"`
	SuperuserUsername string `json:"superuser_username,omitempty"`
	SuperuserPassword string `json:"superuser_password,omitempty"`
	SuperuserURI      string `json:"superuser_uri,omitempty"`
	SuperuserJDBCURI  string `json:"superuser_jdbcUrl,omitempty"`
}

// Bind returns access credentials for a service instance
func (bkr *Broker) Bind(instanceID string, bindingID string, details brokerapi.BindDetails) (brokerapi.BindingResponse, error) {
	logger := bkr.newLoggingSession("bind", lager.Data{"instanceID": instanceID})
	defer logger.Info("done")

	if err := bkr.assertBindPrecondition(instanceID); err != nil {
		logger.Error("preconditions.error", err)
		return brokerapi.BindingResponse{}, err
	}

	cluster, err := bkr.state.LoadCluster(instanceID)
	if err != nil {
		bkr.logger.Error("load-cluster.error", err)
		return brokerapi.BindingResponse{}, fmt.Errorf("Cloud not load cluster %s", instanceID)
	}

	publicPort := cluster.AllocatedPort

	routerHost := bkr.config.BindHost
	appUsername := "dvw7DJgqzFBJC8"
	appPassword := "jkT3TTNebfrh6C"
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres", appUsername, appPassword, routerHost, publicPort)
	// jdbc := fmt.Sprintf("jdbc:postgresql://%s:%d/postgres?username=%s&password=%s", routerHost, publicPort, appUsername, appPassword)
	superuserUsername := "postgres"
	superuserPassword := "Tof2gNVZMz6Dun"
	superuserURI := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres", superuserUsername, superuserPassword, routerHost, publicPort)
	// superuserJDBCURI := fmt.Sprintf("jdbc:postgresql://%s:%d/postgres?username=%s&password=%s", routerHost, publicPort, superuserUsername, superuserPassword)
	return brokerapi.BindingResponse{
		Credentials: CredentialsHash{
			Host:     routerHost,
			Port:     publicPort,
			Username: appUsername,
			Password: appPassword,
			URI:      uri,
			// JDBCURI:           jdbc,
			SuperuserUsername: superuserUsername,
			SuperuserPassword: superuserPassword,
			SuperuserURI:      superuserURI,
			// SuperuserJDBCURI:  superuserJDBCURI,
		},
	}, nil
}

func (bkr *Broker) assertBindPrecondition(instanceID string) error {
	if bkr.state.ClusterExists(instanceID) == false {
		return fmt.Errorf("Service instance %s doesn't exist", instanceID)
	}
	return nil
}
