package licensecheck

import "github.com/pivotal-golang/lager"

func (lc *LicenseCheck) DumpQuotaToLogs() {
	for _, service := range lc.Config.Catalog.Services {
		for _, plan := range service.Plans {
			licenseQuota := lc.TrialQuota(service.ID, plan.ID)
			licenseStatus := "trial"
			if lc.Config.LicenseDetails != nil {
				for _, licensePlan := range lc.Config.LicenseDetails.Plans {
					if licensePlan.UUID == plan.ID {
						licenseQuota = licensePlan.Quota
						licenseStatus = "licensed"
					}
				}
			}
			servicePlanUsage, err := lc.ServicePlanUsage(plan.ID)
			if err != nil {
				lc.Logger.Error("quota-status.cf-lookup", err)
				licenseStatus = "cf-unavailable"
				servicePlanUsage = -1
			}
			data := lager.Data{
				"service-id":    service.ID,
				"plan-id":       plan.ID,
				"status":        licenseStatus,
				"quota":         licenseQuota,
				"current-usage": servicePlanUsage,
			}
			lc.Logger.Info("quota-status", data)
		}
	}
}

func (lc *LicenseCheck) TrialQuota(serviceID, planID string) int {
	return 10
}
