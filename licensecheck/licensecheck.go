package licensecheck

import (
	"github.com/dingotiles/dingo-postgresql-broker/backend"
	"github.com/dingotiles/dingo-postgresql-broker/bkrconfig"
	"github.com/pivotal-golang/lager"
)

// LicenseCheck allows testing of current usage of service broker against CF
type LicenseCheck struct {
	Config *bkrconfig.Config
	etcd   backend.EtcdClient
	Logger lager.Logger
}

// NewLicenseCheck creates LicenseCheck
func NewLicenseCheck(etcd backend.EtcdClient, config *bkrconfig.Config, baseLogger lager.Logger) (lc *LicenseCheck) {
	return &LicenseCheck{
		Config: config,
		etcd:   etcd,
		Logger: baseLogger.Session("license-check"),
	}
}
