# set/existing-empty-table

Set inserts repeated missing values into an existing empty TOML table.

## Source Files

```toml
title = "demo"

[tui.keymap]

[tui]
theme = "light"
```

## Commands

```shell
config set tui.keymap.composer.submit tab
config set tui.keymap.composer.queue alt-q
config set tui.keymap.editor.insert_newline enter
```

## Result Files

```toml
title = "demo"

[tui.keymap]
composer.submit = "tab"
composer.queue = "alt-q"
editor.insert_newline = "enter"

[tui]
theme = "light"
```
