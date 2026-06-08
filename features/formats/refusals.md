# formats/refusals

Unknown-extension files are refused when the format cannot be selected safely.

## Source Files

### TOML (ambiguous.conf)

```toml
port = 3000
```

### YAML (unknown.conf)

```yaml
not config
```

### YAML (bad-hint.conf)

```yaml
# format: xml
server:
  port: 3000
```

## Commands

```shell
config list
exit -> 1
toml/ambiguous.conf !-> ERROR ambiguous config format for ambiguous.conf; add # format: toml or # format: ini
yaml/unknown.conf !-> ERROR cannot determine config format for unknown.conf; add # format: toml, # format: ini, or # format: yaml
yaml/bad-hint.conf !-> ERROR unsupported format hint "xml" for bad-hint.conf
```
