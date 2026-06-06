# unset/inline-mapping

Unset removes fields from inline mappings without expanding them.

## Source Files

```yaml
title: inline mapping demo

database:
  host: localhost
  pool: { min: 1, max: 10 }
```

```toml
title = "inline mapping demo"

[database]
host = "localhost"
pool = { min = 1, max = 10 }
```

## Commands

```shell
config unset database.pool.max
```

## Result Files

```yaml
title: inline mapping demo

database:
  host: localhost
  pool: { min: 1 }
```

```toml
title = "inline mapping demo"

[database]
host = "localhost"
pool = { min = 1 }
```
