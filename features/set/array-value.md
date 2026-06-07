# set/array-index

Update scalar array items by index.

## Source Files

### YAML

```yaml
title: yaml
methods: [GET, SET]
```

### TOML

```toml
title = "toml"
methods = ["GET", "SET"]
```

### JSON

```json
{
  "title": "json",
  "methods": ["GET", "SET"]
}
```

## Commands

```shell
config set methods.1 POST
```

## Result Files

### YAML

```yaml
title: yaml
methods: [GET, POST]
```

### TOML

```toml
title = "toml"
methods = ["GET", "POST"]
```

### JSON

```json
{
  "methods": [
    "GET",
    "POST"
  ],
  "title": "json"
}
```
