# Development & Release Guide

This guide details how to develop, test, and release new versions of **avm**.

---

## 1. Local Development

### Prerequisites
- [Go](https://go.dev/) 1.21 or higher
- [Node.js](https://nodejs.org/) (for plugin development)
- [pnpm](https://pnpm.io/)

### Building the Binary
To build the `avm-bin` binary locally:
```bash
go build -o avm-bin main.go
```

### Testing Plugins Locally
To test a plugin you are developing, use an absolute path to symlink it:
```bash
./avm-bin plugin add $(pwd)/plugins/node
```

### Debugging
If you encounter issues, enable debug logging to see plugin stderr and timeout warnings:
```bash
AVM_DEBUG=1 avm list
```

---

## 2. Release Process

The release process is automated via GitHub Actions. Both **npm** and **Homebrew** are updated whenever a new tag is pushed.

### Step 1: Update Version
Update the version in `package.json`:
```json
{
  "name": "@dracorunner/avm",
  "version": "0.3.0",
  ...
}
```

### Step 2: Commit and Push
```bash
git add package.json
git commit -m "chore: release v0.3.0"
git push origin main
```

### Step 3: Tag and Trigger Release
Create and push a new git tag. This triggers the `Release` (GoReleaser/Brew) and `NPM Publish` workflows.
```bash
git tag -a v0.3.0 -m "Release v0.3.0"
git push origin v0.3.0
```

---

## 3. Automation Details

### Homebrew (GoReleaser)
The `.goreleaser.yml` configuration builds binaries for:
- macOS (amd64, arm64)
- Linux (amd64, arm64)

It automatically updates the formula in the [DracoRunner/homebrew-tap](https://github.com/DracoRunner/homebrew-tap) repository.

### npm
The `npm-publish.yml` workflow installs dependencies and publishes the package to the npm registry under the `@dracorunner` scope.
