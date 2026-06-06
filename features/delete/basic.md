# delete/basic

Delete an entire record collection.

## Source Files

```yaml
title: demo

servers:
  - name: api
    port: 3000
  - name: worker
    port: 3001

owner: ops
```

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

```yaml
title: demo

owner: ops
```

```toml
title = "demo"

[style]
color = "blue"
```
