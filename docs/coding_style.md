# Coding Style Guide

This document describes the coding style guidelines for the project. All code should follow these rules. We use golangci-lint to automatically check and enforce these rules.

## Installing golangci-lint

```bash
make lint-install
```

## Running lint checks

```bash
make lint
```

## Code Style Rules

### 1. Line Length Limit

Each line of code should not exceed 110 characters. This helps improve code readability.

```go
// Not recommended
func someFunction() error {
    if err := someVeryLongFunctionCall(withManyParameters, andEvenMoreParameters, soManyThatItMakesTheLineVeryLong, andHardToRead); err != nil {
        return err
    }
    return nil
}

// Recommended
func someFunction() error {
    if err := someVeryLongFunctionCall(
        withManyParameters,
        andEvenMoreParameters,
        soManyThatItMakesTheLineVeryLong,
        andHardToRead,
    ); err != nil {
        return err
    }
    return nil
}
```

### 2. Line Breaks with Commas

When breaking lines, each element should be on its own line. This is enforced through code reviews and examples rather than automated tools.

```go
// Not recommended
var slice = []string{
    "first", "second",
    "third", "fourth",
}

// Recommended
var slice = []string{
    "first",
    "second",
    "third",
    "fourth",
}
```

### 3. Import Grouping

Import statements should be grouped in the following order:
1. Standard library
2. Project packages (github.com/carv-protocol/d.a.t.a/...)
3. Third-party packages

```go
// Recommended
import (
    "context"
    "fmt"
    "time"
    
    "github.com/carv-protocol/d.a.t.a/src/pkg/database"
    "github.com/carv-protocol/d.a.t.a/src/pkg/logger"
    
    "github.com/google/uuid"
    "go.uber.org/zap"
)
```

You can use the `gci` tool to automatically format imports:

```bash
make gci
```

Or directly:

```bash
gci write --skip-generated -s standard -s "prefix(github.com/carv-protocol/d.a.t.a)" -s default path/to/file.go
```

### 4. Avoid Nested Loops

Try to avoid using nested loops, as they increase code complexity and reduce readability. If nested loops are necessary, consider extracting the inner loop into a separate function.

```go
// Not recommended
func processData(data [][]int) {
    for i := 0; i < len(data); i++ {
        for j := 0; j < len(data[i]); j++ {
            // Process data[i][j]
        }
    }
}

// Recommended
func processData(data [][]int) {
    for i := 0; i < len(data); i++ {
        processRow(data[i])
    }
}

func processRow(row []int) {
    for j := 0; j < len(row); j++ {
        // Process row[j]
    }
}
```

The linter enforces a maximum cyclomatic complexity of 15 and a maximum nesting depth of 5 for conditional statements.

### 5. Avoid Long Functions

Functions should not exceed 80 lines or 50 statements. Long functions should be broken down into smaller functions, each doing one thing.

```go
// Not recommended
func doEverything() {
    // 100+ lines of code doing multiple things
}

// Recommended
func doEverything() {
    doFirstThing()
    doSecondThing()
    doThirdThing()
}

func doFirstThing() {
    // 20-30 lines of code doing one thing
}

func doSecondThing() {
    // 20-30 lines of code doing one thing
}

func doThirdThing() {
    // 20-30 lines of code doing one thing
}
```

### 6. Comments in English

All code comments must be written in English. Comments should be clear, concise, and end with a period.

```go
// Not recommended
// This is a non-English comment

// Recommended
// This function processes user data.
```