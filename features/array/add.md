# array/add

Array add appends one or more scalar values, creates the array when missing,
and ignores values that are already present.

## Source Files

```yaml
roots: ["$HOME/.cache"]
```

```toml
roots = ["$HOME/.cache"]
```

## Commands

```shell
config array add roots /tmp "$HOME/.cache"
config array add extra /var/tmp
```

## Result Files

```yaml
roots: [$HOME/.cache, /tmp]
extra:
  - /var/tmp
```

```toml
roots = ["$HOME/.cache", "/tmp"]
extra = ["/var/tmp"]
```
