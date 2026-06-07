# set/inline-mapping

Set updates and inserts fields inside inline mappings without expanding them.

## Source Files

### YAML

```yaml
title: inline mapping demo

database:
  host: localhost
  pool: { min: 1, max: 10 }
```

### TOML

```toml
title = "inline mapping demo"

[database]
host = "localhost"
pool = { min = 1, max = 10 }
```

## Commands

```shell
config set database.pool.min 2
config set database.pool.default 10
```

## Result Files

### YAML

```yaml
title: inline mapping demo

database:
  host: localhost
  pool: { min: 2, max: 10, default: 10 }
```

### TOML

```toml
title = "inline mapping demo"

[database]
host = "localhost"
pool = { min = 2, max = 10, default = 10 }
```
