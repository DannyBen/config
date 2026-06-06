# unset/if-value

Unset can remove a key only when the current scalar value still matches the
caller-provided value.

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
```

## Result Files

```yaml
queue: alt-w
```

```toml
queue = "alt-w"
```
