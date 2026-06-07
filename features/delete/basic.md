# delete/basic

Delete an entire record collection.

## Source Files

### YAML

```yaml
title: demo

servers:
  - name: api
    port: 3000
  - name: worker
    port: 3001

owner: ops
```

### TOML

```toml
title = "demo"

[[servers]]
name = "api"
port = 3000

[[servers]]
name = "worker"
port = 3001

[style]
color = "blue"
```

## Commands

```shell
config delete servers
```

## Result Files

### YAML

```yaml
title: demo

owner: ops
```

### TOML

```toml
title = "demo"

[style]
color = "blue"
```
