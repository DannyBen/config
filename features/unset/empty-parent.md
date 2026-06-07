# unset/empty-parent

Unset removes scalar fields. YAML preserves the empty parent mapping because the
empty mapping remains visible data. TOML prunes the empty parent table because
table headers are structural containers for following assignments.

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
config unset server.port
```

## Result Files

### YAML

```yaml
title: demo

server:
```

### TOML

```toml
title = "demo"
```

### JSON

```json
{
  "server": {},
  "title": "demo"
}
```
