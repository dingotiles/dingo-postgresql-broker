package licensecheck

import "fmt"

type QuotaStatus struct {
	Quota    int
	Usage    int
	Services []*ServiceQuotaStatus
}

type ServiceQuotaStatus struct {
	GUID          string
	Quota         int
	Usage         int
	LicenseStatus string
	Plans         []*PlanQuotaStatus
}

type PlanQuotaStatus struct {
	GUID          string
	Quota         int
	Usage         int
	LicenseStatus string
}

func (lc *LicenseCheck) FetchQuotaStatus() (quotaStatus *QuotaStatus) {
	quotaStatus = &QuotaStatus{}
	for _, service := range lc.Config.Catalog.Services {
		serviceStatus := &ServiceQuotaStatus{
			LicenseStatus: "trial",
		}
		for _, plan := range service.Plans {
			planStatus := &PlanQuotaStatus{
				LicenseStatus: "trial",
				Quota:         lc.TrialQuota(service.ID, plan.ID),
			}
			if lc.Config.LicenseDetails != nil {
				for _, licensePlan := range lc.Config.LicenseDetails.Plans {
					if licensePlan.UUID == plan.ID {
						planStatus.LicenseStatus = "licensed"
						planStatus.Quota = licensePlan.Quota
					}
				}
			}
			var err error
			planStatus.Usage, err = lc.ServicePlanUsage(plan.ID)
			if err != nil {
				lc.Logger.Error("quota-status.cf-lookup", err)
				planStatus.LicenseStatus = "cf-unavailable"
			}
			serviceStatus.Usage = serviceStatus.Usage + planStatus.Usage
			serviceStatus.Quota = serviceStatus.Quota + planStatus.Quota
			serviceStatus.Plans = append(serviceStatus.Plans, planStatus)
		}
		quotaStatus.Usage = quotaStatus.Usage + serviceStatus.Usage
		quotaStatus.Quota = quotaStatus.Quota + serviceStatus.Quota
		quotaStatus.Services = append(quotaStatus.Services, serviceStatus)
	}
	return
}

func (lc *LicenseCheck) DisplayQuotaStatus() {
	quotaStatus := lc.FetchQuotaStatus()
	fmt.Printf("License status - Total Usage: %d, Total Quota: %d\n", quotaStatus.Usage, quotaStatus.Quota)
}

func (lc *LicenseCheck) TrialQuota(serviceID, planID string) int {
	return 10
}
