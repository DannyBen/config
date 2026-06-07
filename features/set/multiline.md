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

### JSON

```json
{
  "title": "demo app",
  "message": "hello\nworld"
}
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

### JSON

```json
{
  "message": "short",
  "title": "demo app"
}
```
