package broker

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/bkrconfig"
	"github.com/dingotiles/dingo-postgresql-broker/licensecheck"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Broker is the core struct for the Broker webapp
type Broker struct {
	Config       *bkrconfig.Config
	EtcdClient   backend.EtcdClient
	Backends     []bkrconfig.Backend
	LicenseCheck *licensecheck.LicenseCheck

	Logger lager.Logger
}

// NewBroker is a constructor for a Broker webapp struct
func NewBroker(etcdClient backend.EtcdClient, config *bkrconfig.Config) (bkr *Broker) {
	bkr = &Broker{EtcdClient: etcdClient, Config: config}

	bkr.Logger = bkr.setupLogger()

	bkr.LicenseCheck = licensecheck.NewLicenseCheck(bkr.Config, bkr.Logger)
	bkr.LicenseCheck.DisplayQuotaStatus()

	return
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

func (bkr *Broker) setupLogger() lager.Logger {
	logger := lager.NewLogger("dingo-postgresql-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	return logger
}
