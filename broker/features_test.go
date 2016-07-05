package broker

import (
	"fmt"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/testutil"
	"github.com/frodenas/brokerapi"
)

func TestFeatures_FromProvisionDetails_Default(t *testing.T) {
	t.Parallel()

	testPrefix := "TestFeatures_FromProvisionDetails"
	logger := testutil.NewTestLogger(testPrefix, t)
	bkr := &Broker{logger: logger}

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
	bkr := &Broker{logger: logger}

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
