# dump/basic

Dump config data as YAML.

## Source Files

### YAML

```yaml
title: config
style:
  color: red
  size: 14
```

### TOML

```toml
title = "config"

[style]
color = "red"
size = 14
```

### JSON

```json
{
  "title": "config",
  "style": {
    "color": "red",
    "size": 14
  }
}
```

## Commands

```shell
config dump
-> style:
->   color: red
->   size: 14
-> title: config

config dump style
-> color: red
-> size: 14
```
