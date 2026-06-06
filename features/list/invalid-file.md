# list/invalid-file

List refuses invalid input files.

## Source Files

```yaml
title: invalid YAML demo
style:
  font: arial
  broken: [arial
```

```toml
title = "invalid TOML demo"

[style]
font = "arial"
font = "arial"
```

## Commands

```shell
config list
yaml !-> ERROR invalid YAML: yaml: line 3: did not find expected ',' or ']'
toml !-> ERROR invalid TOML: toml: key font is already defined
exit -> 1
```
