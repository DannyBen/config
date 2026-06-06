# unset/empty-parent

Unset removes scalar fields without pruning empty parent containers.

## Source Files

```yaml
server:
  port: 3000
```

```toml
[server]
port = 3000
```

## Commands

```shell
config unset server.port
```

## Result Files

```yaml
server:
```

```toml
[server]
```
