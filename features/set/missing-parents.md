# set/missing-parents

Set creates missing parent containers for nested paths.

## Source Files

```yaml
title: demo

server:
  port: 3000
```

```toml
title = "demo"

[server]
port = 3000
```

## Commands

```shell
config set server.host localhost
config set features.experimental true
```

## Result Files

```yaml
title: demo

server:
  port: 3000
  host: localhost
features:
  experimental: true
```

```toml
title = "demo"

[server]
port = 3000
host = "localhost"

[features]
experimental = true
```
