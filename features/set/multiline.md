# set/multiline

Set can replace an existing multiline string value.

## Source Files

### YAML

```yaml
title: demo app
message: |-
  hello
  world
```

### TOML

```toml
title = "demo app"
message = """hello
world"""
```

## Commands

```shell
config set message short
```

## Result Files

### YAML

```yaml
title: demo app
message: |-
  short
```

### TOML

```toml
title = "demo app"
message = "short"
```
