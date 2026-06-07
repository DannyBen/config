# dump/json

Dump config data as JSON.

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
config dump --json
-> {
->   "style": {
->     "color": "red",
->     "size": 14
->   },
->   "title": "config"
-> }

config dump style --json
-> {
->   "color": "red",
->   "size": 14
-> }
```
