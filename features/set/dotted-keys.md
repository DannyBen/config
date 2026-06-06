# set/dotted-keys

Set uses double dots to address literal dots in keys.

## Source Files

```yaml
network:
  name: public
```

```toml
[network]
name = "public"
```

## Commands

```shell
config set network.public..port 3000
```

## Result Files

```yaml
network:
  name: public
  public.port: 3000
```

```toml
[network]
name = "public"
"public.port" = 3000
```
