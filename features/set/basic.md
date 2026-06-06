# set/basic

Basic set operations

## Source Files

```yaml
title: demo app
```

```toml
title = "demo app"
```

## Commands

```shell
config set title "update works"
config set version "insert works"
config set server.port "nesting works"
```

## Result Files

```yaml
title: update works
version: insert works
server:
  port: nesting works
```

```toml
title = "update works"
version = "insert works"

[server]
port = "nesting works"
```
