# Features

This folder describes what `config` should do.

Each Markdown file is meant to be readable as a small product example: it shows
the starting config file, the command to run, the output a person should see,
and the final config file when the command edits data.

The same files are also the acceptance tests. Running `op acceptance` executes
the command transcripts and checks the results for every format shown in the
feature file.

## Development

Under `Source Files` and `Result Files`, the `###` heading identifies the file.
Code fence languages are only for Markdown highlighting.

Config file headings use the supported format names:

````markdown
### TOML

```toml
title = "config"
```
````

This writes `config.toml` and runs the command transcript with `CONFIG_FILE`
set to that file. `YAML`, `JSON`, and `INI` work the same way.

Use an explicit filename in parentheses when the test needs a different config
file name:

````markdown
### TOML (settings.conf)

```toml
title = "config"
```
````

Other headings create fixture files without changing `CONFIG_FILE`:

````markdown
### value.txt

```text
line one
line two
```
````

Example commands should normally rely on the runner-provided `CONFIG_FILE`
instead of passing `-f` or `--file`.

Files whose names start with `PENDING-`, or files that contain a leading
`> PENDING ...` note before the first section, are documented future behavior
and are skipped by the acceptance runner.
