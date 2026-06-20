# delete/if-exists

Delete can ignore missing containers and missing selected records when the
caller is cleaning up optional config state.

## Source Files

### YAML

```yaml
title: delete if-exists demo

servers:
  - name: api
    port: 3000
  - name: worker
    port: 3001

style:
  color: blue
```

### TOML

```toml
title = "delete if-exists demo"

[[servers]]
name = "api"
port = 3000

[[servers]]
name = "worker"
port = 3001

[style]
color = "blue"
```

### JSON

```json
{
  "title": "delete if-exists demo",
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
  "style": {
    "color": "blue"
  }
}
```

## Commands

```shell
config delete missing --if-exists
config delete servers --on name:web --if-exists
config delete style --if-exists
```

## Result Files

### YAML

```yaml
title: delete if-exists demo

servers:
  - name: api
    port: 3000
  - name: worker
    port: 3001
```

### TOML

```toml
title = "delete if-exists demo"

[[servers]]
name = "api"
port = 3000

[[servers]]
name = "worker"
port = 3001
```

### JSON

```json
{
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
  "title": "delete if-exists demo"
}
```
