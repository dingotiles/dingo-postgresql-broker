package licensecheck

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/dingotiles/dingo-postgresql-broker/bkrconfig"
	"github.com/pivotal-golang/lager"
)

// LicenseCheck allows testing of current usage of service broker against CF
type LicenseCheck struct {
	Config *bkrconfig.Config
	Logger lager.Logger
}

// NewLicenseCheck creates LicenseCheck
func NewLicenseCheck(config *bkrconfig.Config, baseLogger lager.Logger) (lc *LicenseCheck) {
	return &LicenseCheck{
		Config: config,
		Logger: baseLogger.Session("license-check"),
	}
}

func (lc *LicenseCheck) cfClient() (client *cfclient.Client, err error) {
	if lc.Config.CloudFoundry.API == "" {
		return nil, fmt.Errorf("No Cloud Foundry API configuration provided (cf.api)")
	}
	cfconfig := &cfclient.Config{
		ApiAddress:        lc.Config.CloudFoundry.API,
		Username:          lc.Config.CloudFoundry.Username,
		Password:          lc.Config.CloudFoundry.Password,
		SkipSslValidation: lc.Config.CloudFoundry.SkipSSLValidation,
	}
	return cfclient.NewClient(cfconfig)
}
