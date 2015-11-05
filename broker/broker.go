package broker

import (
	"fmt"
	"net/http"
	"os"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Broker is the core struct for the Broker webapp
type Broker struct {
	Catalog  brokerapi.CatalogResponse
	Backends []Backend

	Logger lager.Logger
}

// NewBroker is a constructor for a Broker webapp struct
func NewBroker() (broker *Broker) {
	broker = &Broker{}
	broker.Logger = lager.NewLogger("patroni-broker")
	broker.Logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	broker.Logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	return broker
}

// Run starts the Martini webapp handler
func (bkr *Broker) Run() {
	username := os.Getenv("BROKER_USERNAME")
	if username == "" {
		username = "starkandwayne"
	}

	password := os.Getenv("BROKER_PASSWORD")
	if password == "" {
		password = "starkandwayne"
	}

	credentials := brokerapi.BrokerCredentials{
		Username: username,
		Password: password,
	}
	fmt.Println(credentials)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	brokerAPI := brokerapi.New(bkr, bkr.Logger, credentials)
	http.Handle("/", brokerAPI)
	bkr.Logger.Fatal("http-listen", http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil))
}
