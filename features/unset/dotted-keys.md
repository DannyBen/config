# unset/dotted-keys

Unset uses double dots to address literal dots in keys.

## Source Files

### YAML

```yaml
network:
  public.port: 3000
  public.host: localhost
```

### TOML

```toml
[network]
"public.port" = 3000
"public.host" = "localhost"
```

### JSON

```json
{
  "network": {
    "public.port": 3000,
    "public.host": "localhost"
  }
}
```

### INI

```ini
[network]
public.port = 3000
public.host = localhost
```

## Commands

```shell
config unset network.public..port
```

## Result Files

### YAML

```yaml
network:
  public.host: localhost
```

### TOML

```toml
[network]
"public.host" = "localhost"
```

### JSON

```json
{
  "network": {
    "public.host": "localhost"
  }
}
```

### INI

```ini
[network]
public.host = localhost
```
