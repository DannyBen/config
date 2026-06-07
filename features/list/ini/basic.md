# list/ini/basic

Basic INI list operations

## Source Files

### INI

```ini
# full-line comments are ignored
title = config
server.port = 3000

[style]
color = red
size = 14

[env.prod]
server = app

[env]
prod.server = web
```

## Commands

```shell
config list
-> title=config
-> server..port=3000
-> style.color=red
-> style.size=14
-> env..prod.server=app
-> env.prod..server=web

config list style
-> style.color=red
-> style.size=14

config list server..port
-> server..port=3000

config list env..prod.server
-> env..prod.server=app

config list env.prod..server
-> env.prod..server=web
```
