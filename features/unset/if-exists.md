# unset/if-exists

Unset can ignore missing keys when the caller is cleaning up optional edits.
Missing keys are treated as an already-clean state.

## Source Files

### YAML

```yaml
submit: tab
queue: alt-q
```

### TOML

```toml
submit = "tab"
queue = "alt-q"
```

## Commands

```shell
config unset submit --if-exists
config unset missing --if-exists
```

## Result Files

### YAML

```yaml
queue: alt-q
```

### TOML

```toml
queue = "alt-q"
```
