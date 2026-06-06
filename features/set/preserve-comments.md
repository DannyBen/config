# set/preserve-comments

Comments are preserved as is

## Source Files

```yaml
# the config
title: demo app   # temporary title
```

```toml
# the config
title = "demo app"  # temporary title
```

## Commands

```shell
config set title "update works"
```

## Result Files

```yaml
# the config
title: update works   # temporary title
```

```toml
# the config
title = "update works"  # temporary title
```
