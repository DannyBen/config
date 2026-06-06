# get/multiline

Get prints multiline string values as their resolved text.

## Source Files

```yaml
title: demo app
message: |-
  hello
  world
literal: |-
  alpha
  beta
```

```toml
title = "demo app"
message = """hello
world"""
literal = '''alpha
beta'''
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
