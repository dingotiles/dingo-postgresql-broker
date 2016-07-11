package broker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/testutil"
	"github.com/pborman/uuid"
)

func TestCallbacks_Configures(t *testing.T) {
	t.Parallel()

	testPrefix := "TestCallbacks_WriteRecreationData"
	logger := testutil.NewTestLogger(testPrefix, t)

	callbacks := NewCallbacks(config.Callbacks{}, logger)
	if want, got := false, callbacks.Configured(); want != got {
		t.Fatalf("Callbacks should not be configures")
	}
}

func TestCallbacks_WriteRecreationData(t *testing.T) {
	t.Parallel()

	testPrefix := "TestCallbacks_WriteRecreationData"
	logger := testutil.NewTestLogger(testPrefix, t)

	testDir, err := ioutil.TempDir("", testPrefix)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.Remove(testDir)
	fileName := fmt.Sprintf("%s/%s", testDir, testPrefix)

	config := config.Callbacks{
		ClusterDataBackup: &config.CallbackCommand{
			Command:   "tee",
			Arguments: []string{fileName},
		},
	}

	recreationData := &structs.ClusterRecreationData{
		InstanceID:       "instanceID",
		OrganizationGUID: "OrganizationGUID",
		PlanID:           "PlanID",
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
		AdminCredentials: structs.PostgresCredentials{
			Username: "pgadmin",
			Password: "pw",
		},
		AllocatedPort: 1234,
	}

	callbacks := NewCallbacks(config, logger)
	callbacks.WriteRecreationData(recreationData)

	rawData, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Could not open file %s, Err: %s", fileName, err)
	}

	writtenData := &structs.ClusterRecreationData{}
	json.Unmarshal(rawData, &writtenData)

	if !reflect.DeepEqual(recreationData, writtenData) {
		t.Fatalf("Written Data doesn't equal original. %v != %v", writtenData, recreationData)
	}
}

func TestCallbacks_RestoreRecreationData(t *testing.T) {
	t.Parallel()

	testPrefix := "TestCallbacks_WriteRecreationData"
	logger := testutil.NewTestLogger(testPrefix, t)

	testDir, err := ioutil.TempDir("", testPrefix)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.Remove(testDir)
	fileName := fmt.Sprintf("%s/%s", testDir, testPrefix)

	instanceID := structs.ClusterID(uuid.New())
	recreationData := &structs.ClusterRecreationData{
		InstanceID:       instanceID,
		OrganizationGUID: "OrganizationGUID",
		PlanID:           "PlanID",
		ServiceID:        "ServiceID",
		SpaceGUID:        "SpaceGUID",
		AdminCredentials: structs.PostgresCredentials{
			Username: "pgadmin",
			Password: "pw",
		},
		AllocatedPort: 1234,
	}

	dataRaw, _ := json.Marshal(recreationData)
	err = ioutil.WriteFile(fileName, dataRaw, os.ModePerm)
	if err != nil {
		t.Fatalf("Could not write file")
	}

	cfg := config.Callbacks{
		ClusterDataRestore: &config.CallbackCommand{
			Command:   "cat",
			Arguments: []string{fileName},
		},
	}

	// test that it retrieves the data from stdout
	callbacks := NewCallbacks(cfg, logger)
	restoredData, err := callbacks.RestoreRecreationData(instanceID)
	if err != nil {
		t.Fatalf("Could not open file %s, Err: %s", fileName, err)
	}

	if !reflect.DeepEqual(recreationData, restoredData) {
		t.Fatalf("Retrieved Data doesn't equal original. %v != %v", restoredData, recreationData)
	}

	// test that it passes instanceID to stdin
	cfg = config.Callbacks{
		ClusterDataRestore: &config.CallbackCommand{
			Command:   "tee",
			Arguments: []string{fileName},
		},
	}
	callbacks = NewCallbacks(cfg, logger)
	_, err = callbacks.RestoreRecreationData(instanceID)

	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Could not open file %s, Err: %s", fileName, err)
	}

	if structs.ClusterID(fileContent) != instanceID {
		t.Fatalf("InstanceID %s was not passed via stdin", instanceID)
	}
}
