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

Run Tests
--------

```
make test
```

Playing
-------

Set `$BROKER_URI` to the API for your service broker.

For local dev deployment (`go run main.go broker`) it is:

```
export BROKER_URI=http://starkandwayne:starkandwayne@localhost:3000
```

For a bosh-lite deployment of [dingo-postgresql-release](https://github.com/dingotiles/dingo-postgresql-release) it is:

```
export BROKER_URI=http://starkandwayne:starkandwayne@10.244.21.2:8889
```

To create a service instance use `-XPUT` to hit the `broker.Provision` behavior:

```
export id=b1
export nodes=2
curl -v -XPUT ${BROKER_URI}/v2/service_instances/$id -d "{\"service_id\": \"beb5973c-e1b2-11e5-a736-c7c0b526363d\", \"plan_id\": \"b96d0936-e423-11e5-accb-93d374e93368\", \"parameters\": {\"node-count\": $nodes}}"
curl -v "${BROKER_URI}/v2/service_instances/$id/service_bindings/test-binding" -XPUT -d '{"service_id": "beb5973c-e1b2-11e5-a736-c7c0b526363d", "plan_id": "b96d0936-e423-11e5-accb-93d374e93368"}'
```

To update an existing service instance use `-XPATCH` to reach the `broker.Update` behavior:

```
export id=b1
export nodes=4
curl -v -XPATCH ${BROKER_URI}/v2/service_instances/$id -d "{\"service_id\": \"beb5973c-e1b2-11e5-a736-c7c0b526363d\", \"plan_id\": \"b96d0936-e423-11e5-accb-93d374e93368\", \"parameters\": {\"node-count\": $nodes}}"
```

### Asynchronous API and polling for status

To create a service instance to emulate the Cloud Foundry Service Broker Asynchronous API, add `"accepts_incomplete": true`:

```
export id=b2
export nodes=2
curl -v -XPUT ${BROKER_URI}/v2/service_instances/${id} -d "{\"accepts_incomplete\": true, \"service_id\": \"beb5973c-e1b2-11e5-a736-c7c0b526363d\", \"plan_id\": \"b96d0936-e423-11e5-accb-93d374e93368\", \"parameters\": {\"node-count\": $nodes}}"
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
curl -v -XPUT ${BROKER_URI}/v2/service_instances/${id} -d "{\"accepts_incomplete\": true, \"service_id\": \"beb5973c-e1b2-11e5-a736-c7c0b526363d\", \"plan_id\": \"b96d0936-e423-11e5-accb-93d374e93368\", \"parameters\": {\"node-count\": $nodes}}"
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

### Discover available cells

```
curl ${BROKER_URI}/admin/cells"
```

This will return JSON that looks like:

```
[{"guid":"10.244.21.7","uri":"http://10.244.21.7","config":{"GUID":"10.244.21.7","AvailabilityZone":"z1","URI":"http://10.244.21.7","Username":"containers","Password":"containers"},"az":"z1"},{"guid":"10.244.22.2","uri":"http://10.244.22.2","config":{"GUID":"10.244.22.2","AvailabilityZone":"z2","URI":"http://10.244.22.2","Username":"containers","Password":"containers"},"az":"z2"}]
```

### Create cluster into specific cells

By default, `cf create-service` will allocate containers/nodes of the cluster to cells/vms from its internal scheduling algorithm. If a new cluster needs to be created into specific cells/vms, then this is possible by passing parameters and using the `/admin/cells` information from above.

If you want a two-node cluster to be created into cells with GUIDs "10.244.22.2" and "10.244.21.7", then provide `cell-guids` to the `create-service` command:

```
cf create-service dingo-postgresql cluster good-cells -c '{"cell-guids": ["10.244.22.2", "10.244.21.7"], "node-count": 2}'
```

The same parameters can be used if growing a cluster with `cf update-service`.
