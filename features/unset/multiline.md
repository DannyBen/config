# unset/multiline

Unset removes multiline string values.

## Source Files

```yaml
title: demo app
obsolete: |-
  remove
  me
message: |-
  hello
  world
```

```toml
title = "demo app"
obsolete = """remove
me"""
message = """hello
world"""
```

## Commands

```shell
config unset obsolete
```

## Result Files

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
