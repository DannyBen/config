# list/invalid-file

List refuses invalid input files.

## Source Files

### YAML

```yaml
title: invalid YAML demo
style:
  font: arial
  broken: [arial
```

### TOML

```toml
title = "invalid TOML demo"

[style]
font = "arial"
font = "arial"
```

### JSON

```json
{
  "title": "invalid JSON demo",
  "style": {
    "font": "arial",
  }
}
```

## Commands

```shell
config list
yaml !-> ERROR invalid YAML: yaml: line 3: did not find expected ',' or ']'
toml !-> ERROR invalid TOML: toml: key font is already defined
json !-> ERROR invalid JSON: invalid character '}' looking for beginning of object key string
exit -> 1
```
