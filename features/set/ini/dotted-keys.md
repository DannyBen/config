# set/ini/dotted-keys

Set INI keys and sections with literal dots.

## Source Files

### INI

```ini
server.port = 3000

[env.prod]
server = app

[env]
prod.server = web
```

## Commands

```shell
config set server..port 3001
config set env..prod.server api
config set env.prod..server worker
config set app..example..com.port 8080
```

## Result Files

### INI

```ini
server.port = 3001

[env.prod]
server = api

[env]
prod.server = worker

[app.example.com]
port = 8080
```
