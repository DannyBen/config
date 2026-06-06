# unset/if-exists

Unset can ignore missing keys when the caller is cleaning up optional edits.

## Source Files

```yaml
submit: tab
queue: alt-q
```

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

```yaml
queue: alt-q
```

```toml
queue = "alt-q"
```
