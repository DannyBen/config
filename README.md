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

### Using a config file

Use `config use FILE` to start a shell for repeated commands on one file.
Run `exit` when finished.

```bash
config use app.toml
config get server.port
config list server
exit
```

<img src="support/vhs/use.gif" width="500">

### Editing values

Change a value while preserving comments and nearby formatting.
The demos share this [example TOML file](support/vhs/app.toml).

```bash
config use app.toml
config set server.port 3000
```

<img src="support/vhs/set.gif" width="500">

### Previewing changes

Use `--diff --color` (`-dc`) to preview an edit before applying it.

```bash
config set server.port 3000 --diff --color
config set server.port 3000
```

<img src="support/vhs/diff.gif" width="500">

### Modifying arrays

Add or remove scalar values in TOML, YAML, and JSON arrays.

```bash
config array add plugins.enabled cache
config array delete plugins.enabled metrics
```

<img src="support/vhs/array.gif" width="500">

### Specifying a file directly

For individual commands, use `-f FILE` instead of opening a shell.

```bash
config get -f app.toml server.port
config set -f app.toml logging.level debug
```

<img src="support/vhs/file.gif" width="500">

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

## Contributing / Support

For issues, questions, suggestions, or contributions,
[open an issue](https://github.com/DannyBen/config/issues).
