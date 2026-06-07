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

projects:
  app:
    owner: platform
  worker:
    owner: operations

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

[projects.app]
owner = "platform"

[style]
color = "blue"

[projects.worker]
owner = "operations"
```

### JSON

```json
{
  "title": "demo",
  "servers": [
    {
      "name": "api",
      "port": 3000
    },
    {
      "name": "worker",
      "port": 3001
    }
  ],
  "projects": {
    "app": {
      "owner": "platform"
    },
    "worker": {
      "owner": "operations"
    }
  },
  "owner": "ops"
}
```

## Commands

```shell
config delete servers
config delete projects
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

### JSON

```json
{
  "owner": "ops",
  "title": "demo"
}
```
