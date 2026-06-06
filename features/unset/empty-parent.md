# unset/empty-parent

Unset removes scalar fields. YAML preserves the empty parent mapping because the
empty mapping remains visible data. TOML prunes the empty parent table because
table headers are structural containers for following assignments.

## Source Files

```yaml
title: demo

server:
  port: 3000
```

```toml
title = "demo"

[server]
port = 3000
```

## Commands

```shell
config unset server.port
```

## Result Files

```yaml
title: demo

server:
```

```toml
title = "demo"
```
