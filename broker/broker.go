package broker

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cloudfoundry-community/patroni-broker/backend"
	"github.com/cloudfoundry-community/patroni-broker/config"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Broker is the core struct for the Broker webapp
type Broker struct {
	Config     *config.Config
	EtcdClient *backend.EtcdClient
	Catalog    brokerapi.CatalogResponse
	Backends   []config.Backend

	Logger lager.Logger
}

// NewBroker is a constructor for a Broker webapp struct
func NewBroker(etcdClient *backend.EtcdClient, config *config.Config) (broker *Broker) {
	broker = &Broker{EtcdClient: etcdClient, Config: config}
	broker.Logger = lager.NewLogger("patroni-broker")
	broker.Logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	broker.Logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	return broker
}

// Run starts the Martini webapp handler
func (bkr *Broker) Run() {
	credentials := brokerapi.BrokerCredentials{
		Username: bkr.Config.Broker.Username,
		Password: bkr.Config.Broker.Password,
	}
	port := bkr.Config.Broker.Port

	brokerAPI := brokerapi.New(bkr, bkr.Logger, credentials)
	http.Handle("/", brokerAPI)
	bkr.Logger.Fatal("http-listen", http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil))
}
