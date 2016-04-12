package license

// LicenseDetails contains the purchasors license information and quotas
type LicenseDetails struct {
	Plans []LicenseServicePlan `json:"service_plans"`
}

// LicenseServicePlan describes the quota limits for a specific service plan UUID/name
type LicenseServicePlan struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Quota int    `json:"quota"`
}

// NewLicenseDetails constructs a new LicenseDetails struct
func NewLicenseDetails() (details *LicenseDetails) {
	return &LicenseDetails{}
}
