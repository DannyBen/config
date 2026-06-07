# array/del

Array del removes one or more scalar values, ignores missing values and missing
arrays, and deletes the key when no values remain.

## Source Files

### YAML

```yaml
roots: ["$HOME/.cache", /tmp]
extra: [/var/tmp]
```

### TOML

```toml
roots = ["$HOME/.cache", "/tmp"]
extra = ["/var/tmp"]
```

### JSON

```json
{
  "roots": ["$HOME/.cache", "/tmp"],
  "extra": ["/var/tmp"]
}
```

## Commands

```shell
config array del roots /tmp /missing
config array del extra /var/tmp
config array del absent /tmp
```

## Result Files

### YAML

```yaml
roots: [$HOME/.cache]
```

### TOML

```toml
roots = ["$HOME/.cache"]
```

### JSON

```json
{
  "roots": [
    "$HOME/.cache"
  ]
}
```
