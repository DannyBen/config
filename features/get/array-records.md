# get/array-records

Get values from arrays of hashes

## Source Files

```yaml
servers:
  - name: api
    host: api.local
    port: 3000
  - name: worker
    host: worker.local
    port: 3001
```

```toml
[[servers]]
name = "api"
host = "api.local"
port = 3000

[[servers]]
name = "worker"
host = "worker.local"
port = 3001
```

## Commands

```shell
config get port --in servers --on name:worker
-> 3001

config get servers.1.host
-> worker.local
```
