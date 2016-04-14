package licensecheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pivotal-golang/lager"
)

type servicePlanUsageResponse struct {
	Count int `json:"total_results"`
}

func (lc *LicenseCheck) ServicePlanUsage(planGUID string) (count int, err error) {
	lc.Logger.Info("service-plan-usage", lager.Data{"cf-api": lc.Config.CloudFoundry.API})
	spuResp := &servicePlanUsageResponse{}

	cfclient, err := lc.cfClient()
	if err != nil {
		return -1, fmt.Errorf("Error constructing CF client: %v", err)
	}
	req := cfclient.NewRequest("GET", fmt.Sprintf("/v2/service_plans/%s/service_instances", planGUID))
	resp, err := cfclient.DoRequest(req)
	if err != nil {
		return -1, fmt.Errorf("Error requesting service plan/service instances: %v", err)
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading service plan/service instances body: %v", err)
	}

	err = json.Unmarshal(resBody, &spuResp)
	if err != nil {
		return -1, fmt.Errorf("Error unmarshaling service plan/service instances %v", err)
	}
	return spuResp.Count, nil
}
