# Setting Up the Linter

This document explains how to set up and use the linter for this project.

## Installing golangci-lint

You can install golangci-lint using the provided Makefile command:

```bash
make lint-install
```

Alternatively, you can install it directly:

```bash
# Using Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Using Homebrew (macOS)
brew install golangci-lint

# Using apt (Debian/Ubuntu)
sudo apt install golangci-lint
```

## Running the Linter

To run the linter on the entire project:

```bash
make lint
```

To run the linter on specific files or directories:

```bash
golangci-lint run ./path/to/directory/...
```

## Setting Up Pre-commit Hook

We provide a pre-commit hook that automatically runs the linter before each commit. To set it up:

1. Copy the pre-commit hook script to your Git hooks directory:

```bash
cp docs/pre-commit-hook.sh .git/hooks/pre-commit
```

2. Make the script executable:

```bash
chmod +x .git/hooks/pre-commit
```

Now, the linter will run automatically before each commit, ensuring that your code follows the project's style guidelines.

## Linter Configuration

The linter is configured in the `.golangci.yml` file at the root of the project. This configuration enforces the following rules:

### Basic Code Quality Checks
- gofmt, goimports, gci: Code formatting and import organization
- govet, errcheck, staticcheck, gosimple, ineffassign, unused: Basic code quality checks

### Code Style Checks
- lll: Line length limit of 110 characters
- funlen: Function length limit (80 lines, 50 statements)
- gocyclo, cyclop: Cyclomatic complexity checks (max 15)
- nestif: Nested conditionals check (max depth 5)
- godot, misspell: Comment formatting and spelling
- whitespace: Whitespace formatting

### Security Checks
- gosec: Security vulnerability checks

### Performance Checks
- bodyclose: Ensures HTTP response bodies are closed
- noctx: Ensures context is passed to functions that need it

### Other Useful Checks
- dupl: Code duplication detection
- goconst: Finds repeated strings that could be constants
- gocritic: Various code improvement suggestions
- errorlint: Error handling best practices
- revive: Comprehensive linting rules
- unconvert: Unnecessary type conversion detection
- unparam: Unused function parameter detection

## Import Formatting with GCI

For import grouping, we use the `gci` tool. If you encounter import formatting issues, you can install and run gci separately:

```bash
# Install gci
go install github.com/daixiang0/gci@latest

# Format imports in a file
gci write --skip-generated -s standard -s "prefix(github.com/carv-protocol/d.a.t.a)" -s default path/to/file.go
```

This will organize imports into three groups:
1. Standard library
2. Project packages (github.com/carv-protocol/d.a.t.a/...)
3. Third-party packages

## Disabling Linter Rules

In some cases, you may need to disable specific linter rules for a particular line or block of code. You can do this using special comments:

```go
// To disable a specific linter for a single line
var someVar = someFunc() // nolint:lll

// To disable multiple linters for a single line
var someVar = someFunc() // nolint:lll,errcheck

// To disable a linter for a block of code
//nolint:funlen
func someLongFunction() {
    // Long function code...
}
```

Use these sparingly and only when absolutely necessary.

## Fixing Common Linter Issues

### Line Length

Break long lines into multiple lines, especially for function calls, struct initializations, and long conditions.

### Import Grouping

If you encounter import grouping issues, use the gci tool as described above.

### Function Length

If a function is too long, break it down into smaller, more focused functions.

### Cyclomatic Complexity

Reduce complexity by:
- Extracting complex conditions into named functions
- Breaking nested loops into separate functions
- Using early returns to reduce nesting

### English Comments

Ensure all comments are written in English and end with a period.

### Code Duplication

If the linter reports code duplication:
- Extract duplicated code into helper functions
- Use generics for type-specific duplications
- Consider if the duplication is actually necessary

### Error Handling

Ensure proper error handling:
- Don't ignore returned errors
- Use appropriate error wrapping
- Check for specific error types correctly

## Troubleshooting

### Linter Not Found

If you get a "command not found" error when running `make lint`, ensure that golangci-lint is installed and in your PATH:

```bash
# Check if golangci-lint is installed
which golangci-lint

# If not found, install it
make lint-install

# Add Go bin directory to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

### Command Path Issues

The Makefile is designed to automatically find and use the correct paths for golangci-lint and gci. It will:
1. Look for the commands in your PATH
2. Fall back to the standard Go bin directory
3. Install the tools if they're not found

If you still encounter issues, you can manually set the paths in the Makefile:

```bash
# In Makefile, replace
GOLINT=/path/to/go/bin/golangci-lint
GCI=/path/to/go/bin/gci
```

You can find the correct paths using:

```bash
which golangci-lint
which gci
```

### Import Formatting Issues

If you're having trouble with import formatting, try running gci directly:

```bash
gci write --skip-generated -s standard -s "prefix(github.com/carv-protocol/d.a.t.a)" -s default path/to/file.go
```

### Compatibility Issues

If you encounter errors related to unsupported configuration options in `.golangci.yml`, you may need to update your golangci-lint version or remove the unsupported options. The configuration is designed to be compatible with recent versions of golangci-lint. 