# set/ini/refusals

INI set refuses ambiguous or unsupported edits.

## Source Files

### INI

```ini
port = 3000
port = 3001

[server]
host = localhost
```

## Commands

```shell
config set port 3002
!-> ERROR port has multiple values
exit -> 1

config set server localhost
!-> ERROR server is a section, not a value
exit -> 1

config set env.prod.server web
!-> ERROR INI paths support only key or section.key
exit -> 1

config set port 3000 --in servers --on name:api
!-> ERROR INI set --in is not supported
exit -> 1
```
