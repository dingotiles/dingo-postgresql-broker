package licensecheck

// ProvisionCheck determines if the quota status allows an additional service instance
// Currently only does a simple check of total usage vs total quota
func (lc *LicenseCheck) CanProvision(serviceGUID, planGUID string) bool {
	quotaStatus := lc.FetchQuotaStatus()
	if quotaStatus.Usage >= quotaStatus.Quota {
		return false
	}
	return true
}
