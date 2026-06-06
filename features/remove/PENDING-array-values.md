# remove/array-values

> PENDING Add remove for scalar array values.

Remove deletes one or more scalar values from an array and deletes the parent
key when the array becomes empty.

## Source Files

```yaml
title: remove demo
sandbox_workspace_write:
  writable_roots: ["$HOME/.cache", /tmp, /var/tmp]
  extra_roots: ["$HOME/.local/state"]
```

```toml
title = "remove demo"

[sandbox_workspace_write]
writable_roots = ["$HOME/.cache", "/tmp", "/var/tmp"]
extra_roots = ["$HOME/.local/state"]
```

## Commands

```shell
config remove sandbox_workspace_write.writable_roots /tmp /missing
config remove sandbox_workspace_write.extra_roots "$HOME/.local/state"
```

## Result Files

```yaml
title: remove demo
sandbox_workspace_write:
  writable_roots: ["$HOME/.cache", /var/tmp]
```

```toml
title = "remove demo"

[sandbox_workspace_write]
writable_roots = ["$HOME/.cache", "/var/tmp"]
```
