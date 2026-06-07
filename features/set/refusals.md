# set/refusals

Refuse replacing mappings with scalar values.

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
config set database 3000
yaml !-> ERROR database is a container, not a scalar value
toml !-> ERROR database is a table, not a value
exit -> 1
```
