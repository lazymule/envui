# Contributing to envui

Thanks for your interest in contributing to envui!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/envui.git`
3. Create a feature branch: `git checkout -b my-feature`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes
7. Push to your fork and open a Pull Request

## Development Setup

### Requirements

- Go 1.21 or later
- For Linux: `libgl1-mesa-dev xorg-dev`

### Building

```bash
go build -o envui .
```

### Running Tests

```bash
go test ./...
```

### Code Style

- Run `gofmt` before committing
- Run `go vet ./...` to check for issues
- Keep functions focused and small
- Add comments for exported functions

## Pull Request Guidelines

- Keep PRs focused on a single change
- Update documentation if needed
- Add tests for new functionality
- Ensure all tests pass
- Follow existing code style

## Reporting Bugs

Open an issue with:

- Steps to reproduce
- Expected behavior
- Actual behavior
- Go version and OS

## Feature Requests

Open an issue describing:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Code of Conduct

Be respectful and constructive. We're all here to build something useful.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
