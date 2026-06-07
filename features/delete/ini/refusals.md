# delete/ini/refusals

INI delete refuses scalar and unsupported targets.

## Source Files

### INI

```ini
title = config

[server]
host = localhost
port = 3000
```

## Commands

```shell
config delete server.host
!-> ERROR server.host is a value, use unset to remove fields
exit -> 1

config delete missing
!-> ERROR missing is not set
exit -> 1

config delete server --on name:api
!-> ERROR INI delete --on is not supported
exit -> 1
```
