# delete/indexed-record

Delete one record by index.

## Source Files

### YAML

```yaml
title: delete indexed record demo

servers:
  - name: api
    host: api.local
    port: 3000
  - name: worker
    host: worker.local
    port: 3001
  - name: backup
    host: backup.local
    port: 3002
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

[[servers]]
name = "backup"
host = "backup.local"
port = 3002
```

### JSON

```json
{
  "title": "delete indexed record demo",
  "servers": [
    {
      "name": "api",
      "host": "api.local",
      "port": 3000
    },
    {
      "name": "worker",
      "host": "worker.local",
      "port": 3001
    },
    {
      "name": "backup",
      "host": "backup.local",
      "port": 3002
    }
  ]
}
```

## Commands

```shell
config delete servers.1
```

## Result Files

### YAML

```yaml
title: delete indexed record demo

servers:
  - name: api
    host: api.local
    port: 3000
  - name: backup
    host: backup.local
    port: 3002
```

### TOML

```toml
title = "delete array record demo"

[[servers]]
name = "api"
host = "api.local"
port = 3000

[[servers]]
name = "backup"
host = "backup.local"
port = 3002
```

### JSON

```json
{
  "servers": [
    {
      "host": "api.local",
      "name": "api",
      "port": 3000
    },
    {
      "host": "backup.local",
      "name": "backup",
      "port": 3002
    }
  ],
  "title": "delete indexed record demo"
}
```
