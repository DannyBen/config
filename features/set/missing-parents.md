# set/missing-parents

Set creates missing parent containers for nested paths.

## Source Files

### YAML

```yaml
title: demo

server:
  port: 3000
```

### TOML

```toml
title = "demo"

[server]
port = 3000
```

### JSON

```json
{
  "title": "demo",
  "server": {
    "port": 3000
  }
}
```

## Commands

```shell
config set server.host localhost
config set features.experimental true
```

## Result Files

### YAML

```yaml
title: demo

server:
  port: 3000
  host: localhost
features:
  experimental: true
```

### TOML

```toml
title = "demo"

[server]
port = 3000
host = "localhost"

[features]
experimental = true
```

### JSON

```json
{
  "features": {
    "experimental": true
  },
  "server": {
    "host": "localhost",
    "port": 3000
  },
  "title": "demo"
}
```
