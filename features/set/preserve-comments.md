# set/preserve-comments

Comments are preserved as is

## Source Files

### YAML

```yaml
# the config
title: demo app   # temporary title
```

### TOML

```toml
# the config
title = "demo app"  # temporary title
```

## Commands

```shell
config set title "update works"
```

## Result Files

### YAML

```yaml
# the config
title: update works   # temporary title
```

### TOML

```toml
# the config
title = "update works"  # temporary title
```
