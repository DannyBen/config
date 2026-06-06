# unset/basic

Unset scalar fields.

## Source Files

```yaml
title: unset demo

database:
  host: localhost
  port: 5432
  password: secret
```

```toml
title = "unset demo"

[database]
host = "localhost"
port = 5432
password = "secret"
```

## Commands

```shell
config unset database.password
```

## Result Files

```yaml
title: unset demo

database:
  host: localhost
  port: 5432
```

```toml
title = "unset demo"

[database]
host = "localhost"
port = 5432
```
