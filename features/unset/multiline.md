# unset/multiline

Unset removes multiline string values.

## Source Files

### YAML

```yaml
title: demo app
obsolete: |-
  remove
  me
message: |-
  hello
  world
```

### TOML

```toml
title = "demo app"
obsolete = """remove
me"""
message = """hello
world"""
```

### JSON

```json
{
  "title": "demo app",
  "obsolete": "remove\nme",
  "message": "hello\nworld"
}
```

## Commands

```shell
config unset obsolete
```

## Result Files

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
  "message": "hello\nworld",
  "title": "demo app"
}
```
