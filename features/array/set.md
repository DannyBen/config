# array/set

Array set replaces an array with one or more scalar values.

## Source Files

```yaml
roots: [/tmp]
```

```toml
roots = ["/tmp"]
```

## Commands

```shell
config array set roots "$HOME/.cache" /var/tmp
```

## Result Files

```yaml
roots: [$HOME/.cache, /var/tmp]
```

```toml
roots = ["$HOME/.cache", "/var/tmp"]
```
