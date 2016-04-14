package bkrconfig

import (
	"encoding/json"
	"fmt"
)

// LicenseDetails contains the purchasors license information and quotas
type LicenseDetails struct {
	CompanyName string               `json:"company_name"`
	Plans       []LicenseServicePlan `json:"service_plans"`
}

// LicenseServicePlan describes the quota limits for a specific service plan UUID/name
type LicenseServicePlan struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Quota int    `json:"quota"`
}

// NewLicenseDetails constructs a new LicenseDetails struct
func NewLicenseDetailsFromLicenseText(licenseText string) (details *LicenseDetails, err error) {
	if licenseText == "" {
		return nil, fmt.Errorf("No license file provided, entering trial mode.")
	}
	details = &LicenseDetails{}
	err = json.Unmarshal([]byte(licenseText), details)
	if err != nil {
		err = fmt.Errorf("License file could not be loaded, entering trial mode (%s).", err.Error())
		return
	}
	return
}
