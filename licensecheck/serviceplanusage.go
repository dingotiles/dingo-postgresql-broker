package licensecheck

import (
	"fmt"
	"regexp"

	"github.com/pivotal-golang/lager"
)

func (lc *LicenseCheck) ServicePlanUsage(planID string) (count int, err error) {
	lc.Logger.Info("service-plan-usage", lager.Data{"planID": planID})

	resp, err := lc.etcd.Get("/serviceinstances", false, true)

	if err != nil {
		// if the key wasn't found etcd is available but no clusters exist (yet)
		notFoundRegExp, _ := regexp.Compile("Key not found")
		if notFoundRegExp.FindString(err.Error()) == "Key not found" {
			return 0, nil
		}
		return 0, fmt.Errorf("Error loading: %v", err)
	}

	count = 0
	for _, instance := range resp.Node.Nodes {
		for _, n := range instance.Nodes {
			if n.Key == fmt.Sprintf("%s/plan_id", instance.Key) && n.Value == planID {
				count += 1
				break
			}
		}
	}

	return
}
