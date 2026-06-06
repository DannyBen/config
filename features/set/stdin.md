# set/stdin

Set reads a string value from stdin when the value argument is `-`.

## Source Files

```text value.txt
line one
line two
```

```yaml
title: yaml
```

```toml
title = "toml"
```

## Commands

```shell
config set help - < value.txt
```

## Result Files

```yaml
title: yaml
help: |-
  line one
  line two
```

```toml
title = "toml"
help = """line one
line two
"""
```
