# Development Guide

Guide for developers who want to contribute to Walgo or understand its internals.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Building and Testing](#building-and-testing)
- [Code Style and Standards](#code-style-and-standards)
- [Adding New Features](#adding-new-features)
- [Testing Strategy](#testing-strategy)
- [Debugging](#debugging)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)

## Development Setup

### Prerequisites

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **Git** - Version control
- **Make** - Build automation (optional but recommended)
- **Hugo** - For testing
- **golangci-lint** - For linting
- **Your favorite editor** - VS Code recommended

### Initial Setup

```bash
# 1. Fork and clone the repository
git clone https://github.com/yourusername/walgo.git
cd walgo

# 2. Install Go dependencies
go mod download

# 3. Install development tools
make install-tools
# Or manually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 4. Verify setup
make test

# 5. Build the project
make build
```

### IDE Setup

#### VS Code (Recommended)

Install extensions:
- **Go** (`golang.go`) - Official Go extension
- **Go Test Explorer** - Test runner UI
- **YAML** - YAML syntax support

Recommended settings (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.lintOnSave": "workspace",
  "go.lintTool": "golangci-lint",
  "go.testOnSave": false,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true
}
```

#### GoLand / IntelliJ IDEA

1. Open project directory
2. Enable Go modules support
3. Configure golangci-lint external tool
4. Set up run configurations for tests

### Environment Variables

For development:

```bash
# Enable verbose logging
export WALGO_VERBOSE=true

# Use custom config location
export WALGO_CONFIG_PATH=./walgo.dev.yaml

# Set development mode (if implemented)
export WALGO_ENV=development
```

## Project Structure

### Directory Layout

```
walgo/
├── main.go                  # Application entry point
│
├── cmd/                     # CLI commands (Cobra)
│   ├── root.go             # Root command + global flags
│   ├── init.go             # walgo init
│   ├── build.go            # walgo build
│   ├── deploy.go           # walgo deploy
│   ├── deploy_http.go      # walgo deploy-http
│   ├── update.go           # walgo update
│   ├── status.go           # walgo status
│   ├── optimize.go         # walgo optimize
│   ├── import.go           # walgo import-obsidian
│   ├── setup.go            # walgo setup
│   ├── setup_deps.go       # walgo setup-deps
│   ├── doctor.go           # walgo doctor
│   ├── serve.go            # walgo serve
│   ├── domain.go           # walgo domain
│   ├── version.go          # walgo version
│   └── *_test.go           # Command tests
│
├── internal/                # Internal packages (not exported)
│   ├── config/             # Configuration management
│   │   ├── config.go       # Load/save configuration
│   │   ├── types.go        # Config structs
│   │   └── *_test.go
│   │
│   ├── hugo/               # Hugo integration
│   │   ├── hugo.go         # Hugo CLI wrapper
│   │   └── *_test.go
│   │
│   ├── optimizer/          # Asset optimization
│   │   ├── optimizer.go    # Main engine
│   │   ├── html.go         # HTML optimizer
│   │   ├── css.go          # CSS optimizer
│   │   ├── js.go           # JavaScript optimizer
│   │   ├── types.go        # Optimizer types
│   │   └── *_test.go
│   │
│   ├── walrus/             # Walrus integration
│   │   ├── walrus.go       # site-builder wrapper
│   │   └── *_test.go
│   │
│   ├── deployer/           # Deployment abstraction
│   │   ├── deployer.go     # Interface
│   │   ├── http/           # HTTP adapter
│   │   │   ├── adapter.go
│   │   │   └── *_test.go
│   │   └── sitebuilder/    # On-chain adapter
│   │       ├── adapter.go
│   │       └── *_test.go
│   │
│   └── obsidian/           # Obsidian integration
│       ├── obsidian.go     # Vault import
│       └── *_test.go
│
├── tests/                   # Integration tests
├── docs/                    # Documentation
├── examples/                # Example sites
├── .github/                 # GitHub workflows
│   └── workflows/
│       ├── ci.yml          # CI pipeline
│       └── release.yml     # Release automation
│
├── Makefile                 # Build automation
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── .golangci.yml           # Linter configuration
└── README.md               # Project README
```

### Key Files

- **`main.go`** - Entry point, initializes root command
- **`cmd/root.go`** - Root Cobra command, global flags
- **`internal/config/types.go`** - All configuration structures
- **`Makefile`** - Common development tasks

## Building and Testing

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build with specific version
make build VERSION=1.0.0

# Manual build
go build -o walgo main.go

# Build with custom flags
go build -ldflags "-X main.Version=dev" -o walgo main.go
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/optimizer -v

# Run specific test
go test ./internal/optimizer -run TestOptimizeHTML -v

# Run integration tests
go test -tags=integration ./tests/...

# Benchmark tests
go test -bench=. ./internal/optimizer
```

### Test Coverage

```bash
# Generate coverage report
make coverage

# View coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage percentage
go test -cover ./...
```

### Linting

```bash
# Run linter
make lint

# Auto-fix issues
golangci-lint run --fix

# Run specific linters
golangci-lint run --disable-all --enable=errcheck
```

### Formatting

```bash
# Format all code
make fmt

# Or manually
go fmt ./...
gofmt -s -w .
```

## Code Style and Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

Key points:

1. **Naming Conventions**
   ```go
   // Good - exported, descriptive
   func OptimizeHTML(content string) (string, error)

   // Good - unexported, concise
   func parseConfig(path string) (*Config, error)

   // Bad - unclear abbreviation
   func OptHTML(c string) (string, error)
   ```

2. **Error Handling**
   ```go
   // Good - wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to optimize HTML: %w", err)
   }

   // Bad - lose context
   if err != nil {
       return err
   }
   ```

3. **Comments**
   ```go
   // Good - complete sentence, starts with function name
   // OptimizeHTML minifies HTML content by removing whitespace and comments.
   func OptimizeHTML(content string) (string, error)

   // Bad - incomplete, no context
   // optimize html
   func OptimizeHTML(content string) (string, error)
   ```

4. **Package Comments**
   ```go
   // Package optimizer provides HTML, CSS, and JavaScript optimization.
   //
   // The optimizer processes files in a directory and applies various
   // minification techniques to reduce file sizes while preserving functionality.
   package optimizer
   ```

### Cobra Command Structure

```go
// cmd/example.go
package cmd

import "github.com/spf13/cobra"

var exampleCmd = &cobra.Command{
    Use:   "example [args]",
    Short: "Brief description",
    Long:  `Longer description with examples and usage notes.`,
    Args:  cobra.ExactArgs(1),  // Validate argument count
    RunE:  runExample,           // Use RunE for error handling
}

func init() {
    rootCmd.AddCommand(exampleCmd)

    // Add flags
    exampleCmd.Flags().StringP("flag", "f", "default", "Flag description")
}

func runExample(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### Error Messages

```go
// Good - actionable, specific
return fmt.Errorf("failed to read config file %s: %w. Run 'walgo setup' to create one", path, err)

// Bad - vague, not helpful
return fmt.Errorf("error: %w", err)
```

### Testing Conventions

```go
func TestOptimizeHTML(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     string
        wantErr  bool
    }{
        {
            name:  "removes whitespace",
            input: "<html>  <body>  </body>  </html>",
            want:  "<html><body></body></html>",
        },
        {
            name:    "invalid HTML",
            input:   "<html><unclosed>",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := OptimizeHTML(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("OptimizeHTML() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("OptimizeHTML() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Adding New Features

### Adding a New Command

1. **Create command file:**
   ```bash
   touch cmd/newcmd.go
   ```

2. **Implement command:**
   ```go
   // cmd/newcmd.go
   package cmd

   import (
       "github.com/spf13/cobra"
   )

   var newCmd = &cobra.Command{
       Use:   "new [args]",
       Short: "Brief description",
       Long:  `Detailed description`,
       RunE:  runNew,
   }

   func init() {
       rootCmd.AddCommand(newCmd)

       // Add flags
       newCmd.Flags().String("option", "", "Option description")
   }

   func runNew(cmd *cobra.Command, args []string) error {
       // Implementation
       return nil
   }
   ```

3. **Add tests:**
   ```bash
   touch cmd/newcmd_test.go
   ```

4. **Update documentation:**
   - Add to `docs/COMMANDS.md`
   - Update README if major feature

### Adding a New Internal Package

1. **Create package directory:**
   ```bash
   mkdir -p internal/newpkg
   touch internal/newpkg/newpkg.go
   touch internal/newpkg/newpkg_test.go
   ```

2. **Implement package:**
   ```go
   // internal/newpkg/newpkg.go
   package newpkg

   // NewService creates a new service instance.
   func NewService() *Service {
       return &Service{}
   }

   type Service struct {
       // fields
   }

   // DoSomething performs the main operation.
   func (s *Service) DoSomething() error {
       // implementation
       return nil
   }
   ```

3. **Add comprehensive tests:**
   ```go
   // internal/newpkg/newpkg_test.go
   package newpkg

   import "testing"

   func TestNewService(t *testing.T) {
       s := NewService()
       if s == nil {
           t.Fatal("NewService() returned nil")
       }
   }

   func TestDoSomething(t *testing.T) {
       s := NewService()
       err := s.DoSomething()
       if err != nil {
           t.Errorf("DoSomething() error = %v", err)
       }
   }
   ```

### Adding Configuration Options

1. **Update config types:**
   ```go
   // internal/config/types.go
   type WalgoConfig struct {
       // ... existing fields ...

       NewFeature NewFeatureConfig `mapstructure:"newfeature"`
   }

   type NewFeatureConfig struct {
       Enabled bool   `mapstructure:"enabled"`
       Option  string `mapstructure:"option"`
   }
   ```

2. **Update default config:**
   ```go
   // internal/config/config.go
   func DefaultConfig() *WalgoConfig {
       return &WalgoConfig{
           // ... existing defaults ...

           NewFeature: NewFeatureConfig{
               Enabled: true,
               Option:  "default",
           },
       }
   }
   ```

3. **Update documentation:**
   - Add to `docs/CONFIGURATION.md`

## Testing Strategy

### Unit Tests

Test individual functions in isolation:

```go
func TestOptimizeCSS(t *testing.T) {
    // Arrange
    input := "body { color: #ffffff; }"
    expected := "body{color:#fff}"

    // Act
    result, err := OptimizeCSS(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %q, want %q", result, expected)
    }
}
```

### Integration Tests

Test complete workflows:

```go
// tests/integration_test.go
// +build integration

func TestDeploymentWorkflow(t *testing.T) {
    // Setup test site
    tmpDir := t.TempDir()

    // Run init
    cmd := exec.Command("walgo", "init", "test-site")
    cmd.Dir = tmpDir
    err := cmd.Run()
    require.NoError(t, err)

    // Build site
    cmd = exec.Command("walgo", "build")
    cmd.Dir = filepath.Join(tmpDir, "test-site")
    err = cmd.Run()
    require.NoError(t, err)

    // Deploy HTTP
    cmd = exec.Command("walgo", "deploy-http")
    cmd.Dir = filepath.Join(tmpDir, "test-site")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify output contains URL
    assert.Contains(t, string(output), "https://")
}
```

### Mocking External Dependencies

```go
// For testing, make external commands mockable
var execCommand = exec.Command

func PublishSite(dir string) error {
    cmd := execCommand("site-builder", "publish", dir)
    return cmd.Run()
}

// In tests
func TestPublishSite(t *testing.T) {
    // Save original
    oldExecCommand := execCommand
    defer func() { execCommand = oldExecCommand }()

    // Mock
    execCommand = func(name string, args ...string) *exec.Cmd {
        // Return mock command
        return exec.Command("echo", "mocked")
    }

    // Test
    err := PublishSite("/tmp/site")
    assert.NoError(t, err)
}
```

### Test Coverage Goals

- **Overall:** 80%+
- **Critical paths:** 90%+ (deployer, optimizer)
- **Commands:** 70%+ (harder to test CLI)
- **Internal packages:** 85%+

## Debugging

### Debug Logging

```go
// Add debug logging
import "log"

func processFile(path string) error {
    log.Printf("DEBUG: processing file %s", path)
    // ...
}
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a test
dlv test ./internal/optimizer -- -test.run TestOptimizeHTML

# Debug the application
dlv debug . -- deploy --epochs 1

# In delve:
# break main.main
# continue
# print variable
# step
# next
```

### VS Code Debugging

`.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "args": ["-test.run", "${selectedText}"]
    },
    {
      "name": "Debug walgo deploy",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["deploy", "--epochs", "1"]
    }
  ]
}
```

## CI/CD Pipeline

### GitHub Actions Workflow

The CI pipeline (`.github/workflows/ci.yml`) runs on every push and PR:

1. **Lint** - Code quality checks
2. **Test** - Unit and integration tests
3. **Build** - Cross-platform builds
4. **Security Scan** - Vulnerability detection

### Running CI Locally

```bash
# Install act (GitHub Actions locally)
brew install act

# Run CI
act pull_request

# Run specific job
act -j test
```

### Pre-commit Checks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Format
make fmt

# Lint
make lint

# Test
make test

echo "Pre-commit checks passed!"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- **Major** (1.0.0): Breaking changes
- **Minor** (1.1.0): New features, backward compatible
- **Patch** (1.0.1): Bug fixes

### Creating a Release

1. **Update version:**
   ```bash
   # Update version in version.go or relevant files
   vim cmd/version.go
   ```

2. **Update CHANGELOG:**
   ```bash
   # Document changes
   vim CHANGELOG.md
   ```

3. **Commit and tag:**
   ```bash
   git add .
   git commit -m "Release v1.0.0"
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin main
   git push origin v1.0.0
   ```

4. **GitHub Actions automatically:**
   - Builds binaries for all platforms
   - Creates GitHub release
   - Uploads binaries

### Manual Release (if needed)

```bash
# Build for all platforms
make build-all

# Create GitHub release
gh release create v1.0.0 \
  dist/walgo-* \
  --title "Release v1.0.0" \
  --notes "Release notes here"
```

## Contributing Workflow

See [CONTRIBUTING.md](CONTRIBUTING.md) for the complete contribution guide.

Quick summary:

1. Fork repository
2. Create feature branch
3. Make changes
4. Write tests
5. Update documentation
6. Submit pull request

## Additional Resources

- [Architecture Documentation](ARCHITECTURE.md)
- [Contributing Guide](CONTRIBUTING.md)
- [API Reference](API.md) (coming soon)
- [Cobra Documentation](https://cobra.dev/)
- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
