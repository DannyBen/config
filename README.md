# Config

`config` is a CLI for reading and editing configuration files with a
simple `git config`-like interface.

It is built for hand-written config files. Instead of parsing a file into a map
and writing the whole thing back, `config` plans small source edits so comments,
spacing, table style, YAML anchors, and nearby formatting can stay intact.

Current support focuses on TOML and YAML.  
Future support is planned for INI-like, JSON and perhaps other formats.

## Install

### With eget

```bash
eget dannyben/config
```

### With go install

```bash
go install github.com/dannyben/config@latest
```

### From GitHub Releases

Download the archive for your operating system and CPU from the repository's
[Releases page](https://github.com/DannyBen/config/releases), extract it, and
put the `config` binary somewhere on your `PATH`.

## Highlights

- Read scalar values with script-friendly output.
- Set, unset, delete, and list values by dot path.
- Preserve comments and source formatting where possible.
- Infer common value types such as numbers, booleans, nulls, and dates.
- Create missing parent mappings or tables when the edit is clear.
- Edit records selected by field value, such as `--in servers --on name:api`.
- Preview changes with `--dry` or `--diff`.
- Refuse ambiguous edits instead of silently rewriting the file.

## Usage

```bash
config get path/to/config.toml server.port
config set path/to/config.toml server.port 3000
config array add path/to/config.toml sandbox_workspace_write.writable_roots '$HOME/.cache'
config unset path/to/config.toml server.password
config list path/to/config.toml server
```

For repeated edits, set `CONFIG_FILE` once:

```bash
export CONFIG_FILE=~/.codex/config.toml
config get tui.keymap.composer.submit
config set tui.keymap.composer.submit tab
```

Use `--diff` or `--diff --color` (`-dc`) to preview an edit:

```bash
config set config.yaml server.port 3000 -dc
```

Use `--string` when a value should remain text even if it looks like a typed
literal:

```bash
config set config.toml version 1.0 --string
```

Use `--in` and `--on` to update records:

```bash
config set config.toml port 3000 --in servers --on name:api
```

## Feature Specs

The [features](features/) folder contains readable examples that also run as
acceptance tests.
