package broker

import (
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
)

// Provision a new service instance
func (bkr *Broker) Recreate(instanceID string, acceptsIncomplete bool) (resp brokerapi.ProvisioningResponse, async bool, err error) {
	bkr.Logger.Info("recreate.start", lager.Data{})
	callback := bkr.Config.Callbacks.ClusterDataRestore
	if callback == nil {
		err = fmt.Errorf("Broker not configured to support service recreation")
		bkr.Logger.Error("recreate.restore-callback.missing", err, lager.Data{"missing-config": "callbacks.clusterdata_restore"})
	}
	return
}
