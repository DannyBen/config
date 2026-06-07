# get/ini/refusals

INI get refuses missing and ambiguous values.

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
config get missing
!-> ERROR missing is not set
exit -> 1

config get env.prod.server
!-> ERROR INI paths support only key or section.key
exit -> 1

config get port
!-> ERROR port has multiple values
exit -> 1
```
