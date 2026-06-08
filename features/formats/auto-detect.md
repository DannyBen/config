# formats/auto-detect

Unknown-extension config files are detected from their contents.

## Source Files

### TOML (settings.conf)

```toml
[[servers]]
name = "api"
port = 3000
```

### INI (settings.conf)

```ini
[servers]
host = localhost
```

### YAML (settings.conf)

```yaml
servers:
- name: api
  port: 3000
```

### JSON (settings.conf)

```json
{
  "servers": [
    {
      "name": "api",
      "port": 3000
    }
  ]
}
```

## Commands

```shell
config list servers
toml -> servers.0.name=api
toml -> servers.0.port=3000
ini -> servers.host=localhost
yaml -> servers.0.name=api
yaml -> servers.0.port=3000
json -> servers.0.name=api
json -> servers.0.port=3000
```
