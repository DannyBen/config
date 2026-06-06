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
config array set extra /opt /srv
```

## Result Files

```yaml
roots: [$HOME/.cache, /var/tmp]
extra:
  - /opt
  - /srv
```

```toml
roots = ["$HOME/.cache", "/var/tmp"]
extra = ["/opt", "/srv"]
```
