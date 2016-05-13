package licensecheck

import "fmt"

type QuotaStatus struct {
	Quota    int
	Usage    int
	Services []*ServiceQuotaStatus
}

type ServiceQuotaStatus struct {
	ServiceID     string
	Quota         int
	Usage         int
	LicenseStatus string
	Plans         []*PlanQuotaStatus
}

type PlanQuotaStatus struct {
	PlanID        string
	Quota         int
	Usage         int
	LicenseStatus string
}

func (lc *LicenseCheck) FetchQuotaStatus() (quotaStatus *QuotaStatus) {
	quotaStatus = &QuotaStatus{}
	for _, service := range lc.Config.Catalog.Services {
		serviceStatus := &ServiceQuotaStatus{
			ServiceID:     service.ID,
			LicenseStatus: "trial",
		}
		for _, plan := range service.Plans {
			planStatus := &PlanQuotaStatus{
				PlanID:        plan.ID,
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
				lc.Logger.Error("quota-status.error", err)
				planStatus.LicenseStatus = "unknown"
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
