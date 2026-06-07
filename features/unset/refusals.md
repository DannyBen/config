# unset/refusals

Refuse unsetting mappings.

## Source Files

### YAML

```yaml
database:
  port: 5432
```

### TOML

```toml
[database]
port = 5432
```

## Commands

```shell
config unset database
yaml !-> ERROR database is a container, not a scalar value
toml !-> ERROR database is a table, not a value
exit -> 1
```
