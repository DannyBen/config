# append/array-values

> PENDING Add append for scalar array values.

Append adds one or more scalar values to an array and creates the array when it
does not exist.

## Source Files

```yaml
title: append demo
sandbox_workspace_write:
  writable_roots: ["$HOME/.cache"]
```

```toml
title = "append demo"

[sandbox_workspace_write]
writable_roots = ["$HOME/.cache"]
```

## Commands

```shell
config append sandbox_workspace_write.writable_roots /tmp /var/tmp
config append sandbox_workspace_write.writable_roots /tmp
config append sandbox_workspace_write.extra_roots "$HOME/.local/state" "$HOME/.config"
```

## Result Files

```yaml
title: append demo
sandbox_workspace_write:
  writable_roots: ["$HOME/.cache", /tmp, /var/tmp]
  extra_roots: ["$HOME/.local/state", "$HOME/.config"]
```

```toml
title = "append demo"

[sandbox_workspace_write]
writable_roots = ["$HOME/.cache", "/tmp", "/var/tmp"]
extra_roots = ["$HOME/.local/state", "$HOME/.config"]
```
