package licensecheck

import (
	"fmt"

	"github.com/pivotal-golang/lager"
)

func (lc *LicenseCheck) ServicePlanUsage(planId string) (count int, err error) {
	lc.Logger.Info("service-plan-usage", lager.Data{"planId": planId})

	resp, err := lc.etcd.Get("/serviceinstances", false, true)

	if err != nil {
		return 0, fmt.Errorf("Error loading: %v", err)
	}

	count = 0
	for _, instance := range resp.Node.Nodes {
		for _, n := range instance.Nodes {
			if n.Key == fmt.Sprintf("%s/plan_id", instance.Key) && n.Value == planId {
				count += 1
				break
			}
		}
	}

	return
}
