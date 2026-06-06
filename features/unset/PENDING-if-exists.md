# unset/if-exists

> PENDING Add missing-safe unset with --if-exists.

Unset can ignore missing keys when the caller is cleaning up optional edits.

## Source Files

```yaml
title: unset demo
tui:
  keymap:
    composer:
      submit: tab
```

```toml
title = "unset demo"

[tui.keymap.composer]
submit = "tab"
```

## Commands

```shell
config unset tui.keymap.composer.submit --if-exists
config unset tui.keymap.composer.queue --if-exists
```

## Result Files

```yaml
title: unset demo
tui:
  keymap:
    composer:
```

```toml
title = "unset demo"

[tui.keymap.composer]
```
