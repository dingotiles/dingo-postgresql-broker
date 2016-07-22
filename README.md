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

Test Fakes
---------

To test packages that depend on the interfaces in the `github.com/dingotiles/broker/interfaces` package stand-in implementations located in the `github.com/dingotiles/broker/fakes` package are used.
They are generated with the help of `github.com/maxbrunsfeld/counterfeiter`.

To install counterfeiter:
```
go get github.com/maxbrunsfeld/counterfeiter
```

To generate a new fake for an existing interface (eg. `interfaces.Patroni`):
```
$ counterfeiter -o broker/fakes/fake_patroni.go  broker/interfaces/interfaces.go Patroni
Wrote `FakePatroni` to `broker/fakes/fake_patroni.go`
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

### Lookup internal cluster state

```
curl ${BROKER_URI}/admin/service_instances/05b0d96f-4bd6-4fd1-946c-f6f2fa2a00e4
```

The output will look similar to:

```
{
  "instance_id": "05b0d96f-4bd6-4fd1-946c-f6f2fa2a00e4",
  "service_id": "beb5973c-e1b2-11e5-a736-c7c0b526363d",
  "plan_id": "1545e30e-6dc3-11e5-826a-6c4008a663f0",
  "organization_guid": "de07f7f6-fc1e-45fa-b7be-58ee716d3b3d",
  "space_guid": "33ff71ad-f6d5-4c3c-ac33-440dc6aa4c40",
  "admin_credentials": {
    "username": "pgadmin",
    "password": "xAzGVeEjsNaVCY9r"
  },
  "superuser_credentials": {
    "username": "postgres",
    "password": "ayfUijPa1JtN3GIW"
  },
  "app_credentials": {
    "username": "appuser",
    "password": "spZrDBLBj18W5nS0"
  },
  "allocated_port": 30005,
  "nodes": [
    {
      "node_id": "faed6a46-8f70-4ff9-aeba-a82a08c89574",
      "backend_id": "10.244.21.8",
      "plan_id": "1545e30e-6dc3-11e5-826a-6c4008a663f0",
      "service_id": "beb5973c-e1b2-11e5-a736-c7c0b526363d",
    },
    {
      "node_id": "5d584edd-e5d6-4578-87f5-49089f212b1b",
      "backend_id": "10.244.22.3",
      "plan_id": "1545e30e-6dc3-11e5-826a-6c4008a663f0",
      "service_id": "beb5973c-e1b2-11e5-a736-c7c0b526363d",
    }
  ]
}
```

### Discover available cells

```
curl ${BROKER_URI}/admin/cells"
```

This will return JSON that looks like:

```
[{"guid":"10.244.21.7","uri":"http://10.244.21.7","az":"z1","username":"containers","password":"containers"},{"guid":"10.244.22.2","uri":"http://10.244.22.2","az":"z2","username":"containers","password":"containers"}]
```

### Create cluster into specific cells

By default, `cf create-service` will allocate containers/nodes of the cluster to cells/vms from its internal scheduling algorithm. If a new cluster needs to be created into specific cells/vms, then this is possible by passing parameters and using the `/admin/cells` information from above.

If you want a two-node cluster to be created into cells with GUIDs "10.244.22.2" and "10.244.21.7", then provide `cells` to the `create-service` command:

```
cf create-service dingo-postgresql cluster good-cells -c '{"cells": ["10.244.22.2", "10.244.21.7"], "node-count": 2}'
```

The same parameters can be used if growing a cluster with `cf update-service`.

### Move cluster into different cells

As a database grows, it might require dedicated infrastructure or different cells/vms than it was original provisioned into. An administrator can move a cluster into different cells using the `cf update-service` command and the `cells` parameter (as introduced above).

```
cf update-service their-db -c '{"cells": ["10.244.22.3", "10.244.21.8"]}'
```

This process will first expand the cluster adding two new nodes into the `10.244.22.3` and `10.244.21.8` cells, then failing over the current leader to one of the new replica nodes, and then shutting down the original nodes.

This sequence should result in minimal downtime for bound apps. Bound apps may be required to re-create long lived database connections after this operation.
