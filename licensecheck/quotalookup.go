package licensecheck

import "github.com/pivotal-golang/lager"

func (lc *LicenseCheck) DumpQuotaToLogs() {
	data := lager.Data{}
	lc.Logger.Info("quota-status", data)
}
