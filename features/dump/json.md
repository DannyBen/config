# dump/json

Dump config data as JSON.

## Source Files

```yaml
title: config
style:
  color: red
  size: 14
```

```toml
title = "config"

[style]
color = "red"
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

config dump style --json
-> {
->   "color": "red",
->   "size": 14
-> }
```
