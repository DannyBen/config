# get/basic

Basic set operations

## Source Files

### YAML

```yaml
title: demo app

server:
  port: 3000
  enabled: true

tags:
  - api
  - worker

database:
  host: &host localhost
  replica: *host
```

### TOML

```toml
title = "demo app"
tags = ["api", "worker"]

[server]
port = 3000
enabled = true

[database]
host = "localhost"
replica = "localhost"
```

### JSON

```json
{
  "title": "demo app",
  "tags": ["api", "worker"],
  "server": {
    "port": 3000,
    "enabled": true
  },
  "database": {
    "host": "localhost",
    "replica": "localhost"
  }
}
```

### INI

```ini
title = demo app
tags = [api, worker]

[server]
port = 3000
enabled = true

[database]
host = localhost
replica = localhost
```

## Commands

```shell
config get title
-> demo app
config get server.port
-> 3000
config get server.enabled
-> true
config get tags
-> [api, worker]
config get database.replica
-> localhost
```
