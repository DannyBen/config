# unset/if-value

Unset can remove a key only when the current scalar value still matches the
caller-provided value. Missing keys are treated as an already-clean state.

## Source Files

### YAML

```yaml
submit: tab
queue: alt-w
```

### TOML

```toml
submit = "tab"
queue = "alt-w"
```

### JSON

```json
{
  "submit": "tab",
  "queue": "alt-w"
}
```

## Commands

```shell
config unset submit --if tab
config unset queue --if alt-q
config unset missing --if tab
```

## Result Files

### YAML

```yaml
queue: alt-w
```

### TOML

```toml
queue = "alt-w"
```

### JSON

```json
{
  "queue": "alt-w"
}
```
