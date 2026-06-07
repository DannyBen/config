# set/stdin

Set reads a string value from stdin when the value argument is `-`.

## Source Files

### value.txt

```text value.txt
line one
line two
```

### YAML

```yaml
title: yaml
```

### TOML

```toml
title = "toml"
```

## Commands

```shell
config set help - < value.txt
```

## Result Files

### YAML

```yaml
title: yaml
help: |-
  line one
  line two
```

### TOML

```toml
title = "toml"
help = """line one
line two
"""
```
