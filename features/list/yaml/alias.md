# list/yaml/alias

List resolves YAML aliases.

## Source Files

### YAML

```yaml
defaults: &defaults
  host: localhost
  port: 5432

primary: *defaults
```

## Commands

```shell
config list primary
-> primary.host=localhost
-> primary.port=5432
```
