# unset/if-value

Unset can remove a key only when the current scalar value still matches the
caller-provided value. Missing keys are treated as an already-clean state.

## Source Files

```yaml
submit: tab
queue: alt-w
```

```toml
submit = "tab"
queue = "alt-w"
```

## Commands

```shell
config unset submit --if tab
config unset queue --if alt-q
config unset missing --if tab
```

## Result Files

```yaml
queue: alt-w
```

```toml
queue = "alt-w"
```
