# delete/record

Delete one selected record from a record collection.

## Source Files

### YAML

```yaml
servers:
  - name: api
    port: 3000
  - name: worker
    port: 3001
    host: worker.local
```

### TOML

```toml
title = "delete array record demo"

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
config delete servers --on name:worker
```

## Result Files

### YAML

```yaml
servers:
  - name: api
    port: 3000
```

### TOML

```toml
title = "delete array record demo"

[[servers]]
name = "api"
host = "api.local"
port = 3000
```
