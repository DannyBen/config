# dump/ini/basic

Dump INI data as YAML.

## Source Files

### INI

```ini
title = config
server.port = 3000

[style]
color = red
size = 14
```

## Commands

```shell
config dump
-> server.port: "3000"
-> style:
->   color: red
->   size: "14"
-> title: config

config dump style
-> color: red
-> size: "14"

config dump style.color
-> red
```
