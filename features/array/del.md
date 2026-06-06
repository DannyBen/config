# array/del

Array del removes one or more scalar values, ignores missing values and missing
arrays, and deletes the key when no values remain.

## Source Files

```yaml
roots: ["$HOME/.cache", /tmp]
extra: [/var/tmp]
```

```toml
roots = ["$HOME/.cache", "/tmp"]
extra = ["/var/tmp"]
```

## Commands

```shell
config array del roots /tmp /missing
config array del extra /var/tmp
config array del absent /tmp
```

## Result Files

```yaml
roots: [$HOME/.cache]
```

```toml
roots = ["$HOME/.cache"]
```
