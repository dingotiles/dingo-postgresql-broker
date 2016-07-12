# Configuration

## Callbacks

When the broker successfully provisions a new service instance it can call out to a local script with the details of the provisioned cluster.

```yaml
callbacks:
  provision_success:
    cmd: /path/to/script
    args: [some, args]
```

The incoming STDIN JSON looks like:

```json
{
  "instance_id": "5a223c52-efe1-11e5-849c-4bce32261e9b",
  "service_id": "beb5973c-e1b2-11e5-a736-c7c0b526363d",
  "plan_id": "b96d0936-e423-11e5-accb-93d374e93368",
  "organization_guid": "some-org-guid",
  "space_guid": "some-space-guid",
  "admin_credentials": {
    "username": "pgadmin"
    "password": "pgadminpw"
  },
  "superuser_credentials": {
    "username": "postgres"
    "password": "postgrespw"
  },
  "app_credentials": {
    "username": "appuser"
    "password": "appuserpw"
  },
  "allocated_port": "33004"
}
```

The intent of this callback is to allow operators to backup the service instance's details.
