# set/multiline

Set can replace an existing multiline string value.

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
config set message short
```

## Result Files

```yaml
title: demo app
message: |-
  short
```

```toml
title = "demo app"
message = "short"
```
