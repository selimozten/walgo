# Contributing to Walgo

Thank you for your interest in contributing to Walgo! This document outlines the process for contributing to the project and helps ensure a smooth collaboration experience.

## Table of Contents

- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Filing Issues](#filing-issues)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Development Setup](#development-setup)

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/selimozten/walgo.git
   cd walgo
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```
4. **Run tests** to ensure everything works:
   ```bash
   go test ./...
   ```

## How to Contribute

We welcome contributions in many forms:

- 🐛 **Bug reports and fixes**
- ✨ **New features and enhancements**
- 📚 **Documentation improvements**
- 🧪 **Tests and test improvements**
- 🎨 **Code quality improvements**
- 🌐 **Translations**

## Filing Issues

When filing an issue, please include:

### For Bug Reports
- **Clear description** of the problem
- **Steps to reproduce** the issue
- **Expected vs. actual behavior**
- **Environment details** (OS, Go version, Hugo version, etc.)
- **Relevant logs or error messages**

### For Feature Requests
- **Description** of the proposed feature
- **Motivation** - why is this needed?
- **Proposed solution** or approach
- **Alternative solutions** considered
- **Additional context** or examples

## Submitting Pull Requests

1. **Create a new branch** for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Add or update tests** for your changes

4. **Run the test suite**:
   ```bash
   go test ./...
   go test -race ./...
   ```

5. **Update documentation** if needed

6. **Commit your changes** with clear, descriptive messages:
   ```bash
   git commit -m "Add feature: brief description of changes"
   ```

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request** on GitHub

### Pull Request Guidelines

Your PR should:
- **Link to related issues** (e.g., "Fixes #123" or "Addresses #456")
- **Include a clear description** of what the change does
- **Provide testing instructions** for reviewers
- **Update relevant documentation**
- **Pass all existing tests**
- **Follow the established coding style**

## Coding Standards

### Go Code Style
- Follow standard Go formatting: `go fmt ./...`
- Use `golangci-lint` for linting: `golangci-lint run`
- Write clear, descriptive function and variable names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized

### File Organization
- Place new commands in `cmd/` directory
- Internal packages go in `internal/`
- Tests should be in `*_test.go` files
- Integration tests go in the `tests/` directory

### Commit Messages
- Use imperative mood ("Add feature" not "Added feature")
- Keep first line under 72 characters
- Include more details in the body if needed
- Reference issues and PRs where relevant

## Testing Guidelines

### Unit Tests
- Write tests for new functionality
- Aim for good test coverage
- Use table-driven tests where appropriate
- Mock external dependencies

### Integration Tests
- Add integration tests for new commands
- Test the CLI interface end-to-end
- Include tests for error conditions

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test ./internal/config -v

# Run integration tests
go test -tags=integration ./tests/...
```

## Development Setup

### Prerequisites
- Go 1.22 or later
- Hugo (latest version)
- Git

### Useful Commands
```bash
# Build the project
go build -o walgo main.go

# Install development dependencies
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Format code
go fmt ./...

# Update dependencies
go mod tidy
```

### IDE Setup
We recommend using VS Code with the Go extension, or any editor with Go language server support.

## Code Review Process

1. All submissions require review before merging
2. Maintainers will review PRs in a timely manner
3. Address feedback and update your PR as needed
4. Once approved, a maintainer will merge your PR

## Getting Help

- 💬 **Discussions**: Use GitHub Discussions for questions
- 🐛 **Issues**: Use GitHub Issues for bugs and feature requests
- 📧 **Email**: Contact maintainers for security issues

## Recognition

Contributors are recognized in:
- GitHub contributor list
- Release notes for significant contributions
- Special mentions for first-time contributors

Thank you for contributing to Walgo! 🚀 