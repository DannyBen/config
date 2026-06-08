# formats/hint-comment

Format hints disambiguate unknown-extension config files.

## Source Files

### TOML (settings.conf)

```toml
# format: toml
port = 3000
```

### INI (settings.conf)

```ini
# format: ini
port = 3000
```

### YAML (settings.conf)

```yaml
# format: yaml
server:
  port: 3000
```

## Commands

```shell
config list
toml -> port=3000
ini -> port=3000
yaml -> server.port=3000
```
