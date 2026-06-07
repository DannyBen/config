# set/dotted-keys

Set uses double dots to address literal dots in keys.

## Source Files

### YAML

```yaml
network:
  name: public
```

### TOML

```toml
[network]
name = "public"
```

### JSON

```json
{
  "network": {
    "name": "public"
  }
}
```

## Commands

```shell
config set network.public..port 3000
```

## Result Files

### YAML

```yaml
network:
  name: public
  public.port: 3000
```

### TOML

```toml
[network]
name = "public"
"public.port" = 3000
```

### JSON

```json
{
  "network": {
    "name": "public",
    "public.port": 3000
  }
}
```
