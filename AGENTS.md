# Agent Notes

## Project Shape

- `main.go` is the small binary entrypoint and injects `Version`.
- `cmd/` owns CLI parsing, command behavior, help text, and acceptance tests.
- `format/` owns format dispatch and shared document behavior.
- `format/tomldoc/` owns TOML source-preserving edits.
- `format/yamldoc/` owns YAML source-preserving edits.
- `features/` contains human-readable acceptance specs.
- `op.conf` is the user-facing command catalog.

## Useful Commands

The command catalog is in `op.conf`; inspect it for available Opcode commands
and run those commands with `op`.

```bash
go test ./...
op acceptance
op acceptance get/refusals
op check
go run . --help
```

## Editing Rules

Preserve source style by default. Prefer the smallest valid textual edit that
achieves the requested config change, and refuse risky edits instead of
rewriting or normalizing a user-authored config file.

Keep stdout scriptable. Command payloads and diffs go to stdout; errors go to
stderr.
