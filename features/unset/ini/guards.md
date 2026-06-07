# unset/ini/guards

INI unset supports guarded removals.

## Source Files

### INI

```ini
submit = tab
queue = alt-w
port = 3000 # dev

[server]
host = localhost
port = 3000
```

## Commands

```shell
config unset submit --if-exists
config unset missing --if-exists
config unset queue --if alt-q
config unset port --if "3000 # dev"
config unset server.port --if 3000
```

## Result Files

### INI

```ini
queue = alt-w

[server]
host = localhost
```
