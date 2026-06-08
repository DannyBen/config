<div align='center'>
<img src='config.svg' width=300>
</div>

# Config - CLI for Manipulating Config Files

![repocard](https://repocard.dannyben.com/svg/config.svg)

`config` is a CLI for reading and editing configuration files with a
simple `git config`-like interface.

It is built for hand-written config files. Instead of parsing a file into a map
and writing the whole thing back, `config` plans small source edits so comments,
spacing, table style, YAML anchors, and nearby formatting can stay intact.

Current support includes TOML, YAML, JSON, and INI. TOML, YAML, and INI edits
preserve comments and nearby formatting where possible. JSON edits rewrite the
whole document in canonical pretty JSON.

## Install

The simplest option is with `eget`:

```bash
eget dannyben/config
```

Additional installation methods, including Go, GitHub Release archives, and
Linux `.deb`, `.rpm`, and `.apk` packages, are in [INSTALL.md](INSTALL.md).

## Highlights

- Read scalar values with script-friendly output.
- Set, unset, delete, and list values by dot path.
- Replace, add, and remove scalar array values with `config array` in TOML,
  YAML, and JSON files.
- Detect config formats for files with unknown extensions, with explicit
  comment hints for ambiguous TOML/INI files.
- Preserve comments and source formatting where possible.
- Infer common value types such as numbers, booleans, nulls, and dates where
  the file format supports them.
- Create missing parent mappings or tables when the edit is clear.
- Edit TOML, YAML, and JSON records selected by field value, such as
  `--in servers --on name:api`.
- Preview changes with `--dry` or `--diff`.
- Refuse ambiguous edits instead of silently rewriting the file.
- Includes shell completion for config keys in supported shells.

## Usage

Supported config files:

- TOML: `.toml`
- YAML: `.yaml`, `.yml`
- JSON: `.json`
- INI: `.ini`

Run `config help formats` for format-specific behavior.

```bash
config get -f path/to/config.toml server.port
config set -f path/to/config.toml server.port 3000
config array add -f path/to/config.toml sandbox_workspace_write.writable_roots '$HOME/.cache'
config unset -f path/to/config.toml server.password
config list -f path/to/config.toml server
```

For repeated edits, set `CONFIG_FILE` once:

```bash
export CONFIG_FILE=~/.codex/config.toml
config get tui.keymap.composer.submit
config set tui.keymap.composer.submit tab
config edit
```

Use `--diff` or `--diff --color` (`-dc`) to preview an edit:

```bash
config set -f config.yaml server.port 3000 -dc
```

Use `--string` when a value should remain text even if it looks like a typed
literal:

```bash
config set -f config.toml version 1.0 --string
```

Use `--in` and `--on` to update records:

```bash
config set -f config.toml port 3000 --in servers --on name:api
```

## Feature Specs

The [features](features/) folder contains readable examples that also run as
acceptance tests.
