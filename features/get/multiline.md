# get/multiline

Get prints multiline string values as their resolved text.

## Source Files

### YAML

```yaml
title: demo app
message: |-
  hello
  world
literal: |-
  alpha
  beta
```

### TOML

```toml
title = "demo app"
message = """hello
world"""
literal = '''alpha
beta'''
```

### JSON

```json
{
  "title": "demo app",
  "message": "hello\nworld",
  "literal": "alpha\nbeta"
}
```

## Commands

```shell
config get message
-> hello
-> world

config get literal
-> alpha
-> beta
```
