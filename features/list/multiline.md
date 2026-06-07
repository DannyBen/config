# list/multiline

List renders multiline values on one script-friendly line.

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
config list message
-> message=hello world
```
