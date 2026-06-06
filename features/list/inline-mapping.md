# list/inline-mapping

List flattens inline mappings without expanding their source form.

## Source Files

```yaml
title: inline mapping demo

database:
  host: localhost
  pool: {min: 1, max: 10}
```

```toml
title = "inline mapping demo"

[database]
host = "localhost"
pool = { min = 1, max = 10 }
```

## Commands

```shell
config list database.pool
-> database.pool.min=1
-> database.pool.max=10
```
