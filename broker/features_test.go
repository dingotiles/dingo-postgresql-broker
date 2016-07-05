package broker

import (
	"fmt"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
	"github.com/frodenas/brokerapi"
)

func TestFeatures_FromProvisionDetails_Default(t *testing.T) {
	t.Parallel()

	testPrefix := "TestFeatures_FromProvisionDetails"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler := scheduler.NewScheduler(config.Scheduler{
		Backends: []*config.Backend{},
	}, logger)
	bkr := &Broker{logger: logger, scheduler: scheduler}

	details := brokerapi.ProvisionDetails{}
	features, err := bkr.clusterFeaturesFromProvisionDetails(details)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if features.NodeCount != 2 {
		t.Fatalf("features.NodeCount should be 2 by default")
	}
	if len(features.CellGUIDsForNewNodes) != 0 {
		t.Fatalf("features.CellGUIDsForNewNodes should be empty by default")
	}
}

func TestFeatures_FromProvisionDetails_Overrides(t *testing.T) {
	t.Parallel()

	testPrefix := "TestFeatures_FromProvisionDetails"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler := scheduler.NewScheduler(config.Scheduler{
		Backends: []*config.Backend{
			&config.Backend{GUID: "a"},
			&config.Backend{GUID: "b"},
			&config.Backend{GUID: "c"},
			&config.Backend{GUID: "d"},
		},
	}, logger)
	bkr := &Broker{logger: logger, scheduler: scheduler}

	details := brokerapi.ProvisionDetails{
		Parameters: map[string]interface{}{
			"node-count": 3,
			"cell-guids": []string{"a", "b", "c"},
		},
	}
	features, err := bkr.clusterFeaturesFromProvisionDetails(details)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Printf("%#v\n", features)

	if features.NodeCount != 3 {
		t.Fatalf("features.NodeCount should be 3")
	}
	if len(features.CellGUIDsForNewNodes) != 3 {
		t.Fatalf("Should be 3 items in features.CellGUIDsForNewNodes")
	}
}

func TestFeatures_FromProvisionDetails_Error_UnknownCellGUIDs(t *testing.T) {
	t.Parallel()

	testPrefix := "TestFeatures_FromProvisionDetails"
	logger := testutil.NewTestLogger(testPrefix, t)
	scheduler := scheduler.NewScheduler(config.Scheduler{
		Backends: []*config.Backend{},
	}, logger)
	bkr := &Broker{logger: logger, scheduler: scheduler}

	details := brokerapi.ProvisionDetails{
		Parameters: map[string]interface{}{
			"node-count": 3,
			"cell-guids": []string{"a", "b", "c"},
		},
	}
	_, err := bkr.clusterFeaturesFromProvisionDetails(details)
	if err == nil {
		t.Fatalf("Expect 'Cell GUIDs do not match available cells' error")
	}
}
