# array/ini/refusals

INI does not support array operations.

## Source Files

### INI

```ini
roots = /tmp
```

## Commands

```shell
config array set roots /var/tmp
!-> ERROR INI array set is not supported
exit -> 1

config array add roots /var/tmp
!-> ERROR INI array add is not supported
exit -> 1

config array del roots /tmp
!-> ERROR INI array del is not supported
exit -> 1
```
