Changelog
========================================

v0.1.9 - 2026-06-20
----------------------------------------

- Add config keys to shell completion [`c8a8157`](https://github.com/dannyben/config/commit/c8a8157)
- Fix YAML delete verification [`837d3be`](https://github.com/dannyben/config/commit/837d3be)
- Add `delete --if-exists` [`b1d9b98`](https://github.com/dannyben/config/commit/b1d9b98)
- Compare [`v0.1.8..v0.1.9`](https://github.com/dannyben/config/compare/v0.1.8..v0.1.9)


v0.1.8 - 2026-06-08
----------------------------------------

- Add bash completion to Linux packages [`de7da00`](https://github.com/dannyben/config/commit/de7da00)
- Add `use FILE` command to start a subshell with CONFIG_FILE [`a33aa0b`](https://github.com/dannyben/config/commit/a33aa0b)
- Compare [`v0.1.7..v0.1.8`](https://github.com/dannyben/config/compare/v0.1.7..v0.1.8)


v0.1.7 - 2026-06-08
----------------------------------------

- Replace `CONFIG_FILE` positional arg with `--file PATH` [`5347dce`](https://github.com/dannyben/config/commit/5347dce)
- Improve all help texts [`111a735`](https://github.com/dannyben/config/commit/111a735)
- Add format auto detection and format hint comments [`a3b5e83`](https://github.com/dannyben/config/commit/a3b5e83)
- Compare [`v0.1.6..v0.1.7`](https://github.com/dannyben/config/compare/v0.1.6..v0.1.7)


v0.1.6 - 2026-06-08
----------------------------------------

- Add `edit` command [`9b54a06`](https://github.com/dannyben/config/commit/9b54a06)
- Add `list --color` [`8151184`](https://github.com/dannyben/config/commit/8151184)
- Add TOML implicit parent deletion [`1d82782`](https://github.com/dannyben/config/commit/1d82782)
- Add command aliases: delete=del, list=ls [`53a8302`](https://github.com/dannyben/config/commit/53a8302)
- Improve some help texts [`7c3230d`](https://github.com/dannyben/config/commit/7c3230d)
- Compare [`v0.1.5..v0.1.6`](https://github.com/dannyben/config/compare/v0.1.5..v0.1.6)


v0.1.4 - 2026-06-07
----------------------------------------

- Add INI support [`d6d0419`](https://github.com/dannyben/config/commit/d6d0419)
- Compare [`v0.1.3..v0.1.4`](https://github.com/dannyben/config/compare/v0.1.3..v0.1.4)


v0.1.3 - 2026-06-07
----------------------------------------

- Add `dump` command [`678118e`](https://github.com/dannyben/config/commit/678118e)
- Add `dump --json` [`d8f5af7`](https://github.com/dannyben/config/commit/d8f5af7)
- Add shell `completion` command [`c7e4c55`](https://github.com/dannyben/config/commit/c7e4c55)
- Add `help formats` [`24dbefd`](https://github.com/dannyben/config/commit/24dbefd)
- Add JSON support [`5e246d8`](https://github.com/dannyben/config/commit/5e246d8)
- Compare [`v0.1.2..v0.1.3`](https://github.com/dannyben/config/compare/v0.1.2..v0.1.3)


v0.1.2 - 2026-06-07
----------------------------------------

- Add `delete --if-empty` [`b75e766`](https://github.com/dannyben/config/commit/b75e766)
- Make guarded deletes idempotent for missing keys [`c3eddb4`](https://github.com/dannyben/config/commit/c3eddb4)
- Harden TOML set patches against semantic corruption [`07ff6ec`](https://github.com/dannyben/config/commit/07ff6ec)
- Harden TOML writes and prune empty unset tables [`848b09b`](https://github.com/dannyben/config/commit/848b09b)
- Add semantic verification for YAML patches [`d7221a6`](https://github.com/dannyben/config/commit/d7221a6)
- Prefer explicit TOML tables for deep inserts [`395a3d0`](https://github.com/dannyben/config/commit/395a3d0)
- Group inserted TOML tables by family [`4931d20`](https://github.com/dannyben/config/commit/4931d20)
- Add maintainability checks and split large Go files [`e8711f8`](https://github.com/dannyben/config/commit/e8711f8)
- Compare [`v0.1.1..v0.1.2`](https://github.com/dannyben/config/compare/v0.1.1..v0.1.2)


v0.1.1 - 2026-06-06
----------------------------------------

- Add `unset --if VALUE` [`cbbf4e2`](https://github.com/dannyben/config/commit/cbbf4e2)
- Add `unset --if-exists` [`f65b513`](https://github.com/dannyben/config/commit/f65b513)
- Add array commands [`b4264e6`](https://github.com/dannyben/config/commit/b4264e6)
- Split `array` subcommand helps [`619a35d`](https://github.com/dannyben/config/commit/619a35d)
- Make `set` scalar only (in favor of `array` command space) [`fc1e51e`](https://github.com/dannyben/config/commit/fc1e51e)
- Add style-aware YAML array formatting [`36faee3`](https://github.com/dannyben/config/commit/36faee3)
- Compare [`v0.1.0..v0.1.1`](https://github.com/dannyben/config/compare/v0.1.0..v0.1.1)


v0.1.0 - 2026-06-06
----------------------------------------

- Initial release [`cedd7bb`](https://github.com/dannyben/config/commit/cedd7bb)
- Compare [`v0.1.0`](https://github.com/dannyben/config/compare/v0.1.0)


