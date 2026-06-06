# delete/if-empty

Delete can remove a container only when it has no values. Missing containers are
treated as an already-clean state.

## Source Files

```yaml
tui:
  keymap:
    composer:
      submit: tab
    editor:
      insert_newline: enter
```

```toml
tui.keymap.composer.submit = "tab"

[tui.keymap.editor]
insert_newline = "enter"
```

## Commands

```shell
# should not delete
config delete tui.keymap.composer --if-empty
config delete tui.keymap.missing --if-empty

# should delete
config unset tui.keymap.editor.insert_newline --if enter
config delete tui.keymap.editor --if-empty
config delete tui.keymap.editor --if-empty
```

## Result Files

```yaml
tui:
  keymap:
    composer:
      submit: tab
```

```toml
tui.keymap.composer.submit = "tab"
```
