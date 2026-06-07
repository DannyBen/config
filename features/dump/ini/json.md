# dump/ini/json

Dump INI data as JSON.

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
config dump --json
-> {
->   "server.port": "3000",
->   "style": {
->     "color": "red",
->     "size": "14"
->   },
->   "title": "config"
-> }

config dump style --json
-> {
->   "color": "red",
->   "size": "14"
-> }

config dump style.color --json
-> "red"
```
