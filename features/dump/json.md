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

### INI

```ini
title = config

[style]
color = red
size = 14
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
ini -> {
ini ->   "style": {
ini ->     "color": "red",
ini ->     "size": "14"
ini ->   },
ini ->   "title": "config"
ini -> }

config dump style --json
-> {
->   "color": "red",
->   "size": 14
-> }
ini -> {
ini ->   "color": "red",
ini ->   "size": "14"
ini -> }
```
