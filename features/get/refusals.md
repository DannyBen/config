# get/refusals

Fail on invalid key

## Source Files

```yaml
user:
  name: root
  role: admin

database:
  pool: { min: 1, max: 10 }

servers:
- id: web
  port: 3000
- id: api
  port: 4000
```

```toml
[user]
name = "root"
role = "admin"

[database]
pool = { min = 1, max = 10 }

[[servers]]
name = "web"
port = 3000

[[servers]]
name = "api"
port = 4000
```

## Commands

```shell
config get user
yaml !-> ERROR user is a mapping, not a value
toml !-> ERROR user is a table, not a value
exit -> 1

config get database.pool
yaml !-> ERROR database.pool is a mapping, not a value
toml !-> ERROR database.pool is a table, not a value
exit -> 1

config get servers
yaml !-> ERROR servers is a sequence of records, not a value
toml !-> ERROR servers is an array of records, not a value
exit -> 1
```
