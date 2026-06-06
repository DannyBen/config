Changelog
========================================

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


