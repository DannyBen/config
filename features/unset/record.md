# unset/record

Unset a scalar field from a selected record.

## Source Files

### YAML

```yaml
title: unset record demo
servers:
  - name: api
    port: 3000
    host: api.local
  - name: worker
    port: 3001
    host: worker.local
```

### TOML

```toml
title = "unset record demo"

[[servers]]
name = "api"
port = 3000
host = "api.local"

[[servers]]
name = "worker"
port = 3001
host = "worker.local"
```

### JSON

```json
{
  "title": "unset record demo",
  "servers": [
    {
      "name": "api",
      "port": 3000,
      "host": "api.local"
    },
    {
      "name": "worker",
      "port": 3001,
      "host": "worker.local"
    }
  ]
}
```

## Commands

```shell
config unset port --in servers --on name:worker
config unset servers.0.host
```

## Result Files

### YAML

```yaml
title: unset record demo
servers:
  - name: api
    port: 3000
  - name: worker
    host: worker.local
```

### TOML

```toml
title = "unset record demo"

[[servers]]
name = "api"
port = 3000

[[servers]]
name = "worker"
host = "worker.local"
```

### JSON

```json
{
  "servers": [
    {
      "name": "api",
      "port": 3000
    },
    {
      "host": "worker.local",
      "name": "worker"
    }
  ],
  "title": "unset record demo"
}
```
