# unset/ini/refusals

INI unset refuses ambiguous or unsupported edits.

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
config unset missing
!-> ERROR missing is not set
exit -> 1

config unset port
!-> ERROR port has multiple values
exit -> 1

config unset server
!-> ERROR server is a section, not a value
exit -> 1

config unset env.prod.server
!-> ERROR INI paths support only key or section.key
exit -> 1

config unset host --in servers --on name:api
!-> ERROR INI unset --in is not supported
exit -> 1
```
