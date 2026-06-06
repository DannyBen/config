# list/multiline

List renders multiline values on one script-friendly line.

## Source Files

```yaml
title: demo app
message: |-
  hello
  world
```

```toml
title = "demo app"
message = """hello
world"""
```

## Commands

```shell
config list message
-> message=hello world
```
