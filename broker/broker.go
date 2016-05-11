package broker

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/licensecheck"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Broker is the core struct for the Broker webapp
type Broker struct {
	config       *config.Config
	etcdClient   backend.EtcdClient
	licenseCheck *licensecheck.LicenseCheck
	logger       lager.Logger
	scheduler    *scheduler.Scheduler
}

// NewBroker is a constructor for a Broker webapp struct
func NewBroker(etcdClient backend.EtcdClient, config *config.Config) (bkr *Broker) {
	bkr = &Broker{etcdClient: etcdClient, config: config}

	bkr.logger = bkr.setupLogger()
	bkr.scheduler = bkr.setupScheduler()

	bkr.licenseCheck = licensecheck.NewLicenseCheck(bkr.etcdClient, bkr.config, bkr.logger)
	bkr.licenseCheck.DisplayQuotaStatus()

	return
}

// Run starts the Martini webapp handler
func (bkr *Broker) Run() {
	credentials := brokerapi.BrokerCredentials{
		Username: bkr.config.Broker.Username,
		Password: bkr.config.Broker.Password,
	}
	port := bkr.config.Broker.Port

	brokerAPI := brokerapi.New(bkr, bkr.logger, credentials)
	http.Handle("/", brokerAPI)
	bkr.logger.Fatal("http-listen", http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil))
}

func (bkr *Broker) setupLogger() lager.Logger {
	logger := lager.NewLogger("dingo-postgresql-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	return logger
}

func (bkr *Broker) setupScheduler() *scheduler.Scheduler {
	return scheduler.NewScheduler(bkr.config.Scheduler, bkr.logger)
}
