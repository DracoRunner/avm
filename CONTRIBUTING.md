# Contributing to avm

First off, thank you for considering contributing to **avm**! It's people like you that make **avm** such a great tool.

## How Can I Contribute?

### Reporting Bugs
If you find a bug, please search the issue tracker to see if it has already been reported. If not, please open a new issue and include:
* A clear and descriptive title.
* Steps to reproduce the bug.
* Expected and actual behavior.
* Your operating system and `avm` version.

### Suggesting Enhancements
We are always looking for ways to improve **avm**. If you have an idea, please open a feature request issue!

### Pull Requests
1. Fork the repository and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. Ensure the test suite passes.
4. Make sure your code lints.
5. Issue that pull request!

## Local Development

### Prerequisites
- [Go](https://go.dev/) 1.21 or higher
- [Node.js](https://nodejs.org/) (for plugin development)
- [pnpm](https://pnpm.io/)

### Building the Binary
```bash
go build -o avm-bin main.go
```

### Running Tests
```bash
go test ./...
```

### Testing Plugins Locally
```bash
./avm-bin plugin add $(pwd)/plugins/node
```

## Style Guide
- Follow standard Go idioms and `gofmt`.
- Keep commits focused and provide clear commit messages.
