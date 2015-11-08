Patroni Broker
==============

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
/routing
/postgresql-patroni
/serviceinstances
/service
/cluster
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

### `/cluster`

The router uses confd to generate `haproxy.cfg` files. To make it easier for confd to look up the most useful information for `haproxy.cfg` the information is restructured into the `/cluster` KV directory.

```
curl -s ${ETCD_CLUSTER}/v2/keys/cluster/ | jq -r ".node.nodes[].key"
/cluster/f1
```

```
curl -s ${ETCD_CLUSTER}/v2/keys/cluster/f1/leader | jq -r ".node.nodes[]"
{
  "key": "/cluster/f1/leader/member",
  "value": "f16bc34d-c3de-4843-9dc6-b183cbce2238",
  "modifiedIndex": 39081,
  "createdIndex": 39081
}
{
  "key": "/cluster/f1/leader/host",
  "value": "10.244.21.8",
  "modifiedIndex": 39082,
  "createdIndex": 39082
}
{
  "key": "/cluster/f1/leader/port",
  "value": "32768",
  "modifiedIndex": 39083,
  "createdIndex": 39083
}
```

NOTE: this section is only necessary because patroni doesn't put data into the KV store in a format that confd can parse and apply to templates.

### `/serviceinstance`

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
