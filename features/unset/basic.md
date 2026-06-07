# unset/basic

Unset scalar fields.

## Source Files

### YAML

```yaml
title: unset demo

database:
  host: localhost
  port: 5432
  password: secret
```

### TOML

```toml
title = "unset demo"

[database]
host = "localhost"
port = 5432
password = "secret"
```

### JSON

```json
{
  "title": "unset demo",
  "database": {
    "host": "localhost",
    "port": 5432,
    "password": "secret"
  }
}
```

## Commands

```shell
config unset database.password
```

## Result Files

### YAML

```yaml
title: unset demo

database:
  host: localhost
  port: 5432
```

### TOML

```toml
title = "unset demo"

[database]
host = "localhost"
port = 5432
```

### JSON

```json
{
  "database": {
    "host": "localhost",
    "port": 5432
  },
  "title": "unset demo"
}
```
