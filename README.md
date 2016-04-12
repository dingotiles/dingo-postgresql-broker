Patroni Broker
==============

Usage
-----

```
export PATRONI_BROKER_CONFIG=config/bosh-lite.example.yml
go run main.go broker
go run main.go service-status
go run main.go show-cells
```

Playing
-------

Set `$BROKER_URI` to the API for your service broker.

For local dev deployment (`go run main.go broker`) it is:

```
export BROKER_URI=http://starkandwayne:starkandwayne@localhost:3000
```

For the bosh-lite deployment it is:

```
export BROKER_URI=http://starkandwayne:starkandwayne@10.244.21.4:8888
```

To create a service instance use `-XPUT` to hit the `broker.Provision` behavior:

```
export id=b1
export nodes=2
curl -v -XPUT ${BROKER_URI}/v2/service_instances/$id -d "{\"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"
curl -v "${BROKER_URI}/v2/service_instances/$id/service_bindings/test-binding" -XPUT -d '{"service_id": "0f5c1670-6dc3-11e5-bc08-6c4008a663f0", "plan_id": "1545e30e-6dc3-11e5-826a-6c4008a663f0"}'
```

To update an existing service instance use `-XPATCH` to reach the `broker.Update` behavior:

```
export id=b1
export nodes=4
curl -v -XPATCH ${BROKER_URI}/v2/service_instances/$id -d "{\"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"
```

### Asynchronous API and polling for status

To create a service instance to emulate the Cloud Foundry Service Broker Asynchronous API, add `"accepts_incomplete": true`:

```
export id=b2
export nodes=2
curl -v -XPUT ${BROKER_URI}/v2/service_instances/${id} -d "{\"accepts_incomplete\": true, \"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"
```

Then poll for completion:

```
watch curl -sf ${BROKER_URI}/v2/service_instances/${id}/last_operation
```

The output will progress thru:

```
{"state":"in progress","description":"members stopped, stopped"}
{"state":"in progress","description":"master running; replicas stopped"}
{"state":"succeeded","description":"master running; replicas running"}
```

The see the asynchronous status updates for a new 4-node cluster:

```
export id=b3
export nodes=4
curl -v -XPUT ${BROKER_URI}/v2/service_instances/${id} -d "{\"accepts_incomplete\": true, \"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"
watch curl -sf ${BROKER_URI}/v2/service_instances/${id}/last_operation
```

The state sequence might look like:

```
{"state":"in progress","description":"members stopped, stopped, stopped, stopped"}
{"state":"in progress","description":"master running; replicas stopped, stopped, stopped"}
{"state":"in progress","description":"master running; replicas stopped, starting, stopped"}
{"state":"in progress","description":"master running; replicas starting, running, starting"}
{"state":"succeeded","description":"master running; replicas running, running, running"}
```

### Recreate service API

The broker API also supports "recreate service", which is not a formal API used by the Cloud Controller but could be used by administrators directly.

It is the same API endpoint for "create service", but without providing the JSON data for `service_id` and `plan_id`. If these two data fields are missing then "create service" becomes "recreate service".

This behavior requires the service broker to have been backing up the original "create service" credentials/metadata, as they are used to determine what service/plan/parameters/scale is required to be re-created.

To recreate the `b1` cluster above:

```
export id=b1
curl -v -XPUT ${BROKER_URI}/v2/service_instances/$id -d "{}"
```
