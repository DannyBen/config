# set/basic

Basic set operations

## Source Files

```yaml
servers:
  app.example.com:
    port: 3000
  db.example.com:
    port: 5432
```

```toml
[servers."app.example.com"]
port = 3000

[servers."db.example.com"]
port = 5432
```

## Commands

```shell
config list
-> servers.app..example..com.port=3000
-> servers.db..example..com.port=5432
```
