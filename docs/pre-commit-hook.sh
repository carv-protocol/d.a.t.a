#!/bin/bash

# Pre-commit hook to run golangci-lint and ensure code follows style guidelines
# To install: copy this file to .git/hooks/pre-commit and make it executable

set -e

# Get Go path
GOPATH=$(go env GOPATH)
PATH="$PATH:$GOPATH/bin"

# Get the list of staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

# Exit if no Go files are staged
if [ -z "$STAGED_GO_FILES" ]; then
    echo "No Go files staged for commit. Skipping lint checks."
    exit 0
fi

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Check if gci is installed
if ! command -v gci &> /dev/null; then
    echo "gci not found. Installing..."
    go install github.com/daixiang0/gci@latest
fi

# Get project root directory
PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"

echo "Running gofmt on staged Go files..."
gofmt -w $STAGED_GO_FILES

echo "Running gci on staged Go files..."
for file in $STAGED_GO_FILES; do
    gci write --skip-generated -s standard -s "prefix(github.com/carv-protocol/d.a.t.a)" -s default "$file"
done

echo "Running golangci-lint on staged Go files..."
# Run golangci-lint on staged files with increased timeout
golangci-lint run --timeout=5m $STAGED_GO_FILES

# If we got here, lint checks passed
echo "✅ Lint checks passed!"

# Add any files that might have been modified by formatters
git add $STAGED_GO_FILES

echo "✅ Pre-commit checks completed successfully!"
exit 0 