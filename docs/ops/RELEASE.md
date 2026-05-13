# Release and Publishing

`avm` releases publish the same Rust CLI to three places:

- GitHub release assets
- npm package `@prajanova/avm`
- Prajanova Homebrew tap

## Version checklist

1. Update `package.json` version.
2. Update crate versions when the Rust crates are published or versioned together.
3. Add a matching `CHANGELOG.md` section:

```markdown
## [X.Y.Z] - YYYY-MM-DD
```

4. Run:

```bash
npm run changelog:check
cargo test --workspace
```

For release tags, CI also runs:

```bash
bash scripts/check-release-version.sh
```

This ensures `GITHUB_REF_NAME` matches `v<package.json version>` and that `CHANGELOG.md` contains the same version.

5. Create and push tag:

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

## Release assets

The release pipeline creates these archives:

```text
avm_linux_amd64.tar.gz
avm_linux_arm64.tar.gz
avm_darwin_amd64.tar.gz
avm_darwin_arm64.tar.gz
```

Each archive contains the Rust binary named:

```text
avm-bin
```

Installers keep backward compatibility with older archives that contained `avm`.

## Required secrets

| Secret | Purpose |
| --- | --- |
| `NPM_TOKEN` | Publish `@prajanova/avm` to npm. |
| `HOMEBREW_TAP_TOKEN` | Push formula updates to `prajanova/homebrew-tap`. |

## Homebrew formula behavior

The release workflow updates `Formula/avm.rb` in the tap repository. The formula installs `avm-bin` and verifies:

```bash
avm-bin version
```

## npm behavior

The npm package ships the JavaScript wrapper in `bin/avm-bin.js`. During `postinstall`, it downloads the matching GitHub release archive and installs `bin/avm-bin`.
