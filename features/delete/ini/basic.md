# delete/ini/basic

Delete INI sections.

## Source Files

### INI

```ini
title = config

[server]
host = localhost
port = 3000

[database]
host = db
```

## Commands

```shell
config delete server
```

## Result Files

### INI

```ini
title = config

[database]
host = db
```
