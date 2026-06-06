# unset/refusals

Refuse unsetting mappings.

## Source Files

```yaml
database:
  port: 5432
```

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
