# avm npm wrapper

The npm package `@prajanova/avm` installs a small JavaScript wrapper and downloads the matching Rust binary from GitHub releases.

## Installed commands

```bash
avm
avm-bin
```

Both commands execute the same downloaded Rust binary.

## Install

```bash
npm install -g @prajanova/avm
```

After installation, initialize shell integration:

```bash
eval "$(avm shell-init)"
```

## Release artifact contract

The postinstall script downloads one of:

```text
avm_linux_amd64.tar.gz
avm_linux_arm64.tar.gz
avm_darwin_amd64.tar.gz
avm_darwin_arm64.tar.gz
```

Each archive should contain:

```text
avm-bin
```

Older archives containing `avm` are still accepted by the installer.
