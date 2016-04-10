Patroni Broker
==============

Usage
-----

```
export PATRONI_BROKER_CONFIG=config/bosh-lite.example.yml
dingo-postgresql-broker broker
dingo-postgresql-broker service-status
dingo-postgresql-broker show-cells
```

etcd schema
-----------

Currently only etcd is supported as a backend KV store; but conceptually any KV could be supported.

But it is the same KV store that is used by:

-	registrar - docker containers having their host:port bindings announced in the KV store
-	patroni - cluster of patroni processes forming and reforming clusters of postgresql servers
-	router - discover new clusters and allocate a public port, and update routing to point to leader of cluster

The schema of etcd - as modified and used by the various components, including this broker, is:

```
curl -s ${ETCD_CLUSTER}/v2/keys/ | jq -r ".node.nodes[].key"
/service
/postgresql-patroni
/serviceinstances
/routing
```

To see the entire data store, though it may get large and unwieldy and unsociable to get a DB dump in production:

```
curl -s "${ETCD_CLUSTER}/v2/keys/?recursive=true"
```

### `/service`

`/service` is where each patroni node orchestrates itself into clusters of postgresql servers.

```
id=f1; curl -s ${ETCD_CLUSTER}/v2/keys/service/$id/members | jq -r ".node.nodes[].key"
/service/f1/members/f16bc34d-c3de-4843-9dc6-b183cbce2238
/service/f1/members/c65d2e1a-eb6b-401e-ac9b-195f8f942d26
```

Each of these keys are created by one patroni process running inside a container.

```
id=f1; curl -s ${ETCD_CLUSTER}/v2/keys/service/$id/members/f16bc34d-c3de-4843-9dc6-b183cbce2238 | jq -r .node.value | jq .
{
  "role": "master",
  "state": "running",
  "conn_url": "postgres://replicator:replicator@10.244.21.8:32768/postgres",
  "api_url": "http://127.0.0.1:8008/patroni",
  "xlog_location": 83886336
}
```

Note that the `jq -r .node.value` is passed again into `jq .`. This is because patroni stores its information as a JSON object.

### `/postgresql-patroni`

In order for a patroni process, running inside a Docker container, to discover his `host:port` combination it needs to be able to look it up in the KV store.

The `registrar` service runs on each server and automatically advertises all containers' port information.

`/postgresql-patroni` is where registrar documents each container's host:port binding

With one running cluster of 2 nodes:

```
curl -s ${ETCD_CLUSTER}/v2/keys/postgresql-patroni/ | jq -r ".node.nodes[].key"
/postgresql-patroni/0.patroni.patroni1.patroni.bosh:cf-cb71d10a-c84c-455f-9dbf-ff9bd1e1b8db:5432
/postgresql-patroni/1.patroni.patroni1.patroni.bosh:cf-d0cfa70a-12de-441f-94e7-65a64cb583c0:5432
```

The key path is `/<docker-image>/<hostname>:<internal-id>:<internal-port>`.

Each postgresql/patroni container looks itself up to discover its public `host-ip:port`.

```
curl -s ${ETCD_CLUSTER}/v2/keys/postgresql-patroni/0.patroni.patroni1.patroni.bosh:cf-cb71d10a-c84c-455f-9dbf-ff9bd1e1b8db:5432 | jq -r .node.value
10.244.21.8:32768
```

There is currently no cluster-level information in this data structure. Instead, each `cf-UUID` instance id needs to looked up to determine to which cluster it belongs.

### `/routing`

Each service instance/cluster is allocated a public port that is exposed on each router (so it does not matter which router a TCP request is received as it will be supported by the same port).

The allocated port for each service instance is at `/routing/allocation/:instanceid`.

```
curl -s $ETCD_CLUSTER/v2/keys/routing | jq -r ".node.nodes[].key"
/routing/allocation
/routing/nextport

curl -s $ETCD_CLUSTER/v2/keys/routing/allocation | jq -r ".node.nodes[]"
{
  "key": "/routing/allocation/f1",
  "value": "33006",
  "modifiedIndex": 3176,
  "createdIndex": 3176
}
```

That is, the service instance `f1` (normally would be a long UUID string) has the public router port `33006`.

The value of `/routing/nextport` is the next available public port to be assigned to the next new service instance/cluster.

NOTE: the `/routing` section of data is the only "permanent" data in the KV store. The allocation of a public port to each service instance represents the "contract" made with the end user. We cannot change the public port; but we can change where each service instance node/container is run etc.

### `/serviceinstance`

This `dingo-postgresql-broker` documents the assignment of each container/node in a cluster to a backend broker/cell.

```
curl -s ${ETCD_CLUSTER}/v2/keys/serviceinstances/ | jq -r ".node.nodes[].key"
f1
```

```
curl -s ${ETCD_CLUSTER}/v2/keys/serviceinstances/f1/nodes/ | jq -r ".node.nodes[].key"
/serviceinstances/f1/nodes/f16bc34d-c3de-4843-9dc6-b183cbce2238
/serviceinstances/f1/nodes/c65d2e1a-eb6b-401e-ac9b-195f8f942d26
```

```
curl -s ${ETCD_CLUSTER}/v2/keys/serviceinstances/f1/nodes/f16bc34d-c3de-4843-9dc6-b183cbce2238 | jq -r ".node.nodes[]"
{
  "key": "/serviceinstances/f1/nodes/f16bc34d-c3de-4843-9dc6-b183cbce2238/backend",
  "value": "10.244.21.8",
  "modifiedIndex": 39064,
  "createdIndex": 39064
}
```

Playing
-------

Set `$BROKER_URI` to the API for your service broker.

For the bosh-lite deployment it is:

```
export BROKER_URI=http://starkandwayne:starkandwayne@10.244.21.4:8888
```

To create a service instance use `-XPUT` to hit the `broker.Provision` behavior:

```
id=b1; nodes=2; curl -v -XPUT ${BROKER_URI}/v2/service_instances/$id -d "{\"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"; curl -v "${BROKER_URI}/v2/service_instances/$id/service_bindings/test" -XPUT -d '{"service_id": "0f5c1670-6dc3-11e5-bc08-6c4008a663f0", "plan_id": "1545e30e-6dc3-11e5-826a-6c4008a663f0"}'
```

To update an existing service instance use `-XPATCH` to reach the `broker.Update` behavior:

```
id=b1; nodes=4; curl -v -XPATCH ${BROKER_URI}/v2/service_instances/$id -d "{\"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"; curl -v "${BROKER_URI}/v2/service_instances/$id/service_bindings/test" -XPUT -d '{"service_id": "0f5c1670-6dc3-11e5-bc08-6c4008a663f0", "plan_id": "1545e30e-6dc3-11e5-826a-6c4008a663f0"}'
```

To create a service instance to emulate asynchronous API, add `"accepts_incomplete": true`:

```
export id=b2
nodes=2; curl -v -XPUT ${BROKER_URI}/v2/service_instances/${id} -d "{\"accepts_incomplete\": true, \"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"
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

For a 4-node cluster:

```
id=b3; nodes=4; curl -v -XPUT ${BROKER_URI}/v2/service_instances/${id} -d "{\"accepts_incomplete\": true, \"service_id\": \"0f5c1670-6dc3-11e5-bc08-6c4008a663f0\", \"plan_id\": \"1545e30e-6dc3-11e5-826a-6c4008a663f0\", \"parameters\": {\"node-count\": $nodes}}"; watch curl -sf ${BROKER_URI}/v2/service_instances/${id}/last_operation
```

The state sequence might look like:

```
{"state":"in progress","description":"members stopped, stopped, stopped, stopped"}
{"state":"in progress","description":"master running; replicas stopped, stopped, stopped"}
{"state":"in progress","description":"master running; replicas stopped, starting, stopped"}
{"state":"in progress","description":"master running; replicas starting, running, starting"}
{"state":"succeeded","description":"master running; replicas running, running, running"}
```
