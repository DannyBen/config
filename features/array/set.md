# array/set

Array set replaces an array with one or more scalar values.

## Source Files

### YAML

```yaml
roots: [/tmp]
```

### TOML

```toml
roots = ["/tmp"]
```

## Commands

```shell
config array set roots "$HOME/.cache" /var/tmp
config array set extra /opt /srv
```

## Result Files

### YAML

```yaml
roots: [$HOME/.cache, /var/tmp]
extra:
  - /opt
  - /srv
```

### TOML

```toml
roots = ["$HOME/.cache", "/var/tmp"]
extra = ["/opt", "/srv"]
```
