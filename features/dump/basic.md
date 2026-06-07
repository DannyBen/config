# dump/basic

Dump config data as YAML.

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
config dump
-> style:
->   color: red
->   size: 14
-> title: config

config dump style
-> color: red
-> size: 14
```
