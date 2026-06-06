# set/toml-insertion-style

Set uses nearby TOML structure as a style hint when inserting nested paths.

## Source Files

```toml
root = "demo"
service.api.port = 9000

[server]
port = 3000

[env.prod]
port = 80

[env.dev]
port = 3000

[app]
prod.port = 80
```

## Commands

```shell
config set server.host localhost
config set env.debug.port 8080
config set app.dev.port 3001
config set service.api.host localhost
config set cache.redis.port 6379
```

## Result Files

```toml
root = "demo"
service.api.port = 9000
service.api.host = "localhost"

[server]
port = 3000
host = "localhost"

[env.prod]
port = 80

[env.dev]
port = 3000

[app]
prod.port = 80
dev.port = 3001

[env.debug]
port = 8080

[cache.redis]
port = 6379
```
