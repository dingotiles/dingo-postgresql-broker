package licensecheck

import (
	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// LicenseCheck allows testing of current usage of service broker against CF
type LicenseCheck struct {
	catalog        brokerapi.Catalog
	licenseDetails *config.LicenseDetails
	etcd           etcd.KeysAPI
	logger         lager.Logger
}

// NewLicenseCheck creates LicenseCheck
func NewLicenseCheck(config *config.Config, logger lager.Logger) (*LicenseCheck, error) {
	licenseCheck := &LicenseCheck{
		catalog:        config.Catalog,
		licenseDetails: config.LicenseDetails,
		logger:         logger,
	}

	var err error
	licenseCheck.etcd, err = licenseCheck.setupEtcd(config.Etcd)
	if err != nil {
		return nil, err
	}

	return licenseCheck, nil
}

func (r *LicenseCheck) setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}
