# delete/ini/if-empty

Delete empty INI sections.

## Source Files

### INI

```ini
title = config

[server]
host = localhost

[cache]
```

## Commands

```shell
# should not delete
config delete server --if-empty
config delete missing --if-empty

# should delete
config delete cache --if-empty
```

## Result Files

### INI

```ini
title = config

[server]
host = localhost
```
