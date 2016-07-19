package structs

import "testing"

func TestStructs_ClusterFeaturesFromParameters_Defaults(t *testing.T) {
	t.Parallel()

	var emptyParams map[string]interface{}
	features, err := ClusterFeaturesFromParameters(emptyParams)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if features.NodeCount != 2 {
		t.Fatalf("features.NodeCount should be 2 by default")
	}
	if len(features.CellGUIDs) != 0 {
		t.Fatalf("features.CellGUIDs should be empty by default")
	}
}

func TestFeatures_FromProvisionDetails_Overrides(t *testing.T) {
	t.Parallel()

	params := map[string]interface{}{
		"node-count": 3,
		"cells":      []string{"a", "b", "c"},
	}
	features, err := ClusterFeaturesFromParameters(params)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if features.NodeCount != 3 {
		t.Fatalf("features.NodeCount should be 3")
	}
	if len(features.CellGUIDs) != 3 {
		t.Fatalf("Should be 3 items in features.CellGUIDs")
	}
}

func TestFeatures_FromProvisionDetails_Illegal(t *testing.T) {
	t.Parallel()

	params := map[string]interface{}{
		"node-count": -1,
		"cells":      []string{"a", "b", "c"},
	}
	_, err := ClusterFeaturesFromParameters(params)
	if err == nil {
		t.Fatalf("Expected Error on negative input")
	}
}
