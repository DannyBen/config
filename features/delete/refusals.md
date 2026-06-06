# delete/refusals

Refuse unsafe or invalid delete targets.

## Source Files

```yaml
title: delete refusals demo

servers:
  - name: api
    host: api.local
    port: 3000

style:
  color: blue
  font: arial
```

```toml
title = "delete refusals demo"

[[servers]]
name = "api"
host = "api.local"
port = 3000

[style]
color = "blue"
font = "arial"
```

## Commands

```shell
config delete style.color
!-> ERROR style.color is a value, use unset to remove fields
exit -> 1

config delete servers.2
!-> ERROR servers has no record at index 2
exit -> 1

config delete servers --on name:web
!-> ERROR servers has no records matching name:web
exit -> 1
```
