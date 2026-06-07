# list/ini-dotted-keys

List INI keys and sections with literal dots.

## Source Files

### INI

```ini
server.port = 3000

[app.example.com]
port = 3000

[db.example.com]
port = 5432

[servers]
app.example.com.port = 3000
db.example.com.port = 5432
```

## Commands

```shell
config list
-> server..port=3000
-> app..example..com.port=3000
-> db..example..com.port=5432
-> servers.app..example..com..port=3000
-> servers.db..example..com..port=5432

config list app..example..com
-> app..example..com.port=3000

config list servers.app..example..com..port
-> servers.app..example..com..port=3000
```
