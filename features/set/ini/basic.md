# set/ini/basic

Basic INI set operations.

## Source Files

### INI

```ini
title = demo app

[server]
host = localhost
```

## Commands

```shell
config set title "update works"
config set version "insert works"
config set server.port "section insert works"
config set database.host "new section works"
```

## Result Files

### INI

```ini
title = update works
version = insert works

[server]
host = localhost
port = section insert works

[database]
host = new section works
```
