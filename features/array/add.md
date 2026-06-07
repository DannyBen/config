# array/add

Array add appends one or more scalar values, creates the array when missing,
and ignores values that are already present.

## Source Files

### YAML

```yaml
roots: ["$HOME/.cache"]
```

### TOML

```toml
roots = ["$HOME/.cache"]
```

### JSON

```json
{
  "roots": ["$HOME/.cache"]
}
```

## Commands

```shell
config array add roots /tmp "$HOME/.cache"
config array add extra /var/tmp
```

## Result Files

### YAML

```yaml
roots: [$HOME/.cache, /tmp]
extra:
  - /var/tmp
```

### TOML

```toml
roots = ["$HOME/.cache", "/tmp"]
extra = ["/var/tmp"]
```

### JSON

```json
{
  "extra": [
    "/var/tmp"
  ],
  "roots": [
    "$HOME/.cache",
    "/tmp"
  ]
}
```
