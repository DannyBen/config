# dump/ini/refusals

INI dump refuses ambiguous data.

## Source Files

### INI

```ini
port = 3000
port = 3001

server = localhost

[server]
port = 3000
```

## Commands

```shell
config dump port
!-> ERROR port has multiple values
exit -> 1

config dump server
!-> ERROR server matches both a key and section
exit -> 1

config dump
!-> ERROR port has multiple values
exit -> 1
```
