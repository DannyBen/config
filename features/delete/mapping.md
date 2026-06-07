# delete/mapping

Delete a mapping.

## Source Files

### YAML

```yaml
title: delete mapping demo

server:
  port: 3000

style:
  color: blue
  font: arial
```

### TOML

```toml
title = "delete mapping demo"

[server]
port = 3000

[style]
color = "blue"
font = "arial"
```

### JSON

```json
{
  "title": "delete mapping demo",
  "server": {
    "port": 3000
  },
  "style": {
    "color": "blue",
    "font": "arial"
  }
}
```

## Commands

```shell
config delete style
```

## Result Files

### YAML

```yaml
title: delete mapping demo

server:
  port: 3000
```

### TOML

```toml
title = "delete mapping demo"

[server]
port = 3000
```

### JSON

```json
{
  "server": {
    "port": 3000
  },
  "title": "delete mapping demo"
}
```
