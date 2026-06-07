# Install

`config` is distributed as release archives, Linux packages, and a Go module.

The examples below use a version variable. Set it to the release you want to
install:

```bash
VERSION=0.1.5
BASE="https://github.com/DannyBen/config/releases/download/v$VERSION"
```

## With eget

```bash
eget dannyben/config
```

## With go install

```bash
go install github.com/dannyben/config@latest
```

## Debian / Ubuntu

For `amd64`:

```bash
wget "$BASE/config_${VERSION}_amd64.deb"
sudo apt install "./config_${VERSION}_amd64.deb"
```

For `arm64`, use `config_${VERSION}_arm64.deb`.

## Fedora / RHEL

For `x86_64`:

```bash
wget "$BASE/config-${VERSION}-1.x86_64.rpm"
sudo dnf install "./config-${VERSION}-1.x86_64.rpm"
```

For `aarch64`, use `config-${VERSION}-1.aarch64.rpm`.

## Alpine

For `x86_64`:

```bash
wget "$BASE/config_${VERSION}_x86_64.apk"
sudo apk add --allow-untrusted "./config_${VERSION}_x86_64.apk"
```

For `aarch64`, use `config_${VERSION}_aarch64.apk`.

The `--allow-untrusted` flag is needed because the release `.apk` package is
not signed by an Alpine package repository key.

## GitHub Release Archives

Download the archive for your operating system and CPU from the
[Releases page](https://github.com/DannyBen/config/releases), extract it, and
put the `config` binary somewhere on your `PATH`.

For example, on Linux `amd64`:

```bash
wget "$BASE/config-${VERSION}-linux_amd64.tar.gz"
tar -xzf "config-${VERSION}-linux_amd64.tar.gz"
sudo install -m 755 config /usr/local/bin/config
```

Other archive names use the same pattern:

- `config-${VERSION}-linux_arm64.tar.gz`
- `config-${VERSION}-darwin_amd64.tar.gz`
- `config-${VERSION}-darwin_arm64.tar.gz`

## Shell Completion

Shell completion scripts can be generated with `config completion`.

```bash
config completion bash
config completion zsh
config completion fish
```

Linux packages install the `config` binary only. They do not install shell
completion scripts automatically.
