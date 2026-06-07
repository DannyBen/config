# set/basic

Basic set operations

## Source Files

### YAML

```yaml
title: demo app
```

### TOML

```toml
title = "demo app"
```

### JSON

```json
{
  "title": "demo app"
}
```

## Commands

```shell
config set title "update works"
config set version "insert works"
config set server.port "nesting works"
```

## Result Files

### YAML

```yaml
title: update works
version: insert works
server:
  port: nesting works
```

### TOML

```toml
title = "update works"
version = "insert works"

[server]
port = "nesting works"
```

### JSON

```json
{
  "server": {
    "port": "nesting works"
  },
  "title": "update works",
  "version": "insert works"
}
```
