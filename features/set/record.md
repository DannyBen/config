# set/record

Set updates or creates one selected record in a sequence of records.

## Source Files

```yaml
servers:
  - name: api
    port: 3000
  - name: worker
    port: 3001
  - name: backend
    port: 4000
```

```toml
[[servers]]
name = "api"
port = 3000

[[servers]]
name = "worker"
port = 3001

[[servers]]
name = "backend"
port = 4000
```

## Commands

```shell
# update
config set port 3002 --in servers --on name:worker

# insert
config set port 3003 --in servers --on name:cache

# rename
config set name app --in servers --on name:backend

# indexed
config set servers.3.port 5000
```

## Result Files

```yaml
servers:
  - name: api
    port: 3000
  - name: worker
    port: 3002
  - name: app
    port: 4000
  - name: cache
    port: 5000
```

```toml
[[servers]]
name = "api"
port = 3000

[[servers]]
name = "worker"
port = 3002

[[servers]]
name = "app"
port = 4000

[[servers]]
name = "cache"
port = 5000
```
