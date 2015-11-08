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
curl -s http://184.72.129.218:14001/v2/keys/ | jq -r ".node.nodes[].key"
/routing
/postgresql-patroni
/serviceinstances
/service
/cluster
```

### `/postgresql-patroni`

-	`/postgresql-patroni` is where registrar documents each container's host:port binding

With one running cluster of 2 nodes:

```
curl -s http://184.72.129.218:14001/v2/keys/postgresql-patroni/ | jq -r ".node.nodes[].key"
/postgresql-patroni/0.patroni.patroni1.patroni.bosh:cf-cb71d10a-c84c-455f-9dbf-ff9bd1e1b8db:5432
/postgresql-patroni/1.patroni.patroni1.patroni.bosh:cf-d0cfa70a-12de-441f-94e7-65a64cb583c0:5432
```

The key path is `/<docker-image>/<hostname>:<internal-id>:<internal-port>`.

Each postgresql/patroni container looks itself up to discover its public `host-ip:port`.

```
curl -s http://184.72.129.218:14001/v2/keys/postgresql-patroni/0.patroni.patroni1.patroni.bosh:cf-cb71d10a-c84c-455f-9dbf-ff9bd1e1b8db:5432 | jq -r .node.value
10.244.21.6:32775
```
