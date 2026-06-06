# set/structured-looking-strings

Single values that look like structures are stored as strings.

## Source Files

```yaml
title: yaml
```

```toml
title = "toml"
```

## Commands

```shell
config set array-text '[3000, 4000]'
config set mapping-text '{ min = 1, max = 10 }'
```

## Result Files

```yaml
title: yaml
array-text: "[3000, 4000]"
mapping-text: "{ min = 1, max = 10 }"
```

```toml
title = "toml"
array-text = "[3000, 4000]"
mapping-text = "{ min = 1, max = 10 }"
```
