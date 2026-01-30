# Contributors Guide

This guide will help familiarize contributors to the `rawbytedev/zerokv` repository. ZeroKV is a minimal, zero-overhead key-value store abstraction for Go that allows you to switch between different database backends without changing your application code.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Implementing a New Storage Backend](#implementing-a-new-storage-backend)
- [Testing Requirements](#testing-requirements)
- [Using ZeroKV](#using-zerokv)
- [Pull Request Checklist](#pull-request-checklist)

---

## Getting Started

### Prerequisites

- Go 1.25.2 or higher
- Git
- Basic understanding of key-value stores

### Cloning the Repository

```bash
git clone https://github.com/rawbytedev/zerokv.git
cd zerokv
```

### Installing Dependencies

```bash
go mod download
go mod tidy
```

---

## Development Setup

### Project Structure

```text
zerokv/
├── interface.go          # Core interfaces (Core, Iterator, Batch)
├── go.mod              # Module definition
├── badgerdb/           # BadgerDB implementation
│   ├── badgerdb.go     # Main implementation
│   ├── badgerdb_test.go # Implementation-specific tests
│   └── options.go      # Configuration options
├── pebbledb/           # PebbleDB implementation
│   ├── pebbledb.go     # Main implementation
│   ├── pebbledb_test.go # Implementation-specific tests
│   └── options.go      # Configuration options
├── tests/              # Shared integration tests
│   ├── crud_test.go    # CRUD operation tests
│   └── iterator_test.go # Iterator tests
├── helpers/            # Testing utilities
│   ├── test_setups.go  # Test database setup
│   └── context_helpers.go # Context utilities
└── examples/           # Usage examples
    ├── basic_usage.go
    ├── multi_usage.go
    └── runtime_switch_usage.go
```

### Running Tests Locally

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with race detector
go test ./... -race

# Run specific test package
go test ./tests -v
go test ./badgerdb -v
go test ./pebbledb -v
```

### Building the Project

```bash
# Build all packages
go build ./...

# Run code quality checks
go vet ./...
```

---

## Code Style Guidelines

### Naming Conventions

- **Interfaces**: Use descriptive names (e.g., `Core`, `Iterator`, `Batch`)
- **Structs**: Use camelCase, no abbreviated names (e.g., `badgerDB`, not `bdb`)
- **Methods**: Descriptive action verbs (e.g., `Put()`, `Get()`, `Delete()`)
- **Constants**: Use PascalCase (e.g., `PutOp`, `GetOp`)
- **Unexported**: Lowercase first letter (e.g., `badgerDB`, `badgerIterator`)

### File Organization

Each implementation package should follow this structure:

1. **Package declaration and imports**
2. **Type definitions** (structs for Core, Batch, Iterator)
3. **Constructor** (`New<DBName>()`)
4. **Core interface methods** (Put, Get, Delete, Close)
5. **Batch methods** (Put, Delete, Commit)
6. **Iterator methods** (Next, Key, Value, Release, Error)
7. **Special methods** (optional, implementation-specific)

### Documentation Comments

All exported functions and types must have documentation comments:

```go
// Put inserts or updates a key-value pair in the database.
// Returns an error if the operation fails.
func (b *badgerDB) Put(ctx context.Context, key []byte, data []byte) error {
    // Implementation
}

// badgerIterator represents an iterator over BadgerDB key-value pairs.
type badgerIterator struct {
    Iterator *badger.Iterator
    started  bool
    valid    bool
    err      []error
}
```

### Error Handling

- Always check and propagate errors
- Use context cancellation: `if err := ctx.Err(); err != nil { return err }`
- Return meaningful error messages
- Avoid silent failures

### Context Handling

All operations that accept `context.Context` must:

1. Check for context cancellation at the start
2. Respect context deadlines
3. Return context errors appropriately

```go
func (b *badgerDB) Get(ctx context.Context, key []byte) ([]byte, error) {
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    // Implementation
}
```

### Memory Management

- Document resource cleanup requirements
- Use defer for cleanup operations
- Ensure iterators call `Release()` to avoid leaks

```go
defer it.Release()
```

---

## Implementing a New Storage Backend

### 1. Create Package Structure

Create a new directory for your backend:

```bash
mkdir newdb
touch newdb/newdb.go newdb/newdb_test.go newdb/options.go
```

### 2. Define Your Configuration

In `newdb/options.go`:

```go
package newdb

// Config holds configuration for NewDB
type Config struct {
    Dir        string
    // Add backend-specific options
}

func DefaultOptions(dir string) *Config {
    return &Config{Dir: dir}
}
```

### 3. Implement Core Interface

In `newdb/newdb.go`, implement all methods from `zerokv.Core`:

```go
package newdb

import (
    "context"
    "github.com/rawbytedev/zerokv"
    // Import your database library
)

type newDB struct {
    db *YourDBType
}

type newBatch struct {
    batch *YourBatchType
}

type newIterator struct {
    Iterator *YourIteratorType
    started  bool
    valid    bool
    err      []error
}

// NewNewDB initializes and returns a zerokv.Core instance
func NewNewDB(cfg Config) (zerokv.Core, error) {
    // Initialize your database
    db, err := YourDB.Open(cfg.Dir)
    if err != nil {
        return nil, err
    }
    return &newDB{db: db}, nil
}

// Implement Core methods
func (n *newDB) Put(ctx context.Context, key []byte, data []byte) error {
    if err := ctx.Err(); err != nil {
        return err
    }
    return n.db.Set(key, data)
}

func (n *newDB) Get(ctx context.Context, key []byte) ([]byte, error) {
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    return n.db.Get(key)
}

func (n *newDB) Delete(ctx context.Context, key []byte) error {
    if err := ctx.Err(); err != nil {
        return err
    }
    return n.db.Delete(key)
}

func (n *newDB) Close() error {
    return n.db.Close()
}

func (n *newDB) Batch() zerokv.Batch {
    return &newBatch{batch: n.db.NewBatch()}
}

func (n *newDB) Scan(prefix []byte) zerokv.Iterator {
    it := n.db.NewIterator(prefix)
    return &newIterator{Iterator: it, started: false, valid: false}
}

// Implement Batch methods
func (b *newBatch) Put(key []byte, data []byte) error {
    return b.batch.Set(key, data)
}

func (b *newBatch) Delete(key []byte) error {
    return b.batch.Delete(key)
}

func (b *newBatch) Commit(ctx context.Context) error {
    if err := ctx.Err(); err != nil {
        return err
    }
    return b.batch.Write()
}

// Implement Iterator methods
func (it *newIterator) Next() bool {
    if !it.started {
        it.valid = it.Iterator.First()
        it.started = true
    } else {
        it.valid = it.Iterator.Next()
    }
    return it.valid
}

func (it *newIterator) Key() []byte {
    if !it.valid {
        return nil
    }
    return it.Iterator.Key()
}

func (it *newIterator) Value() []byte {
    if !it.valid {
        return nil
    }
    return it.Iterator.Value()
}

func (it *newIterator) Release() {
    it.Iterator.Close()
}

func (it *newIterator) Error() error {
    if len(it.err) == 0 {
        return nil
    }
    return it.err[len(it.err)-1]
}
```

### 4. Critical Requirements

When implementing a new backend, ensure:

- **All interface methods are implemented**
- **Context is checked at the start of each operation**
- **Error handling is consistent across all operations**
- **Iterator.Error() handles empty error slices** (return nil)
- **Iterator.Release() is properly implemented** to avoid leaks
- **Comments document all exported functions**
- **Edge cases are handled** (empty keys, nil values, etc.)

### 5. Add Implementation-Specific Tests

In `newdb/newdb_test.go`:

```go
package newdb_test

import (
    "testing"
    "github.com/rawbytedev/zerokv/helpers"
    "github.com/stretchr/testify/require"
)

func TestNewDBBatchOperations(t *testing.T) {
    db := helpers.SetupDB(t, "newdb")
    batch := db.Batch()
    
    key := helpers.RandomBytes(16)
    value := helpers.RandomBytes(32)
    
    err := batch.Put(key, value)
    require.NoError(t, err)
    
    err = batch.Commit(t.Context())
    require.NoError(t, err)
    
    retrievedValue, err := db.Get(t.Context(), key)
    require.NoError(t, err)
    require.Equal(t, value, retrievedValue)
    
    defer db.Close()
}
```

### 6. Register in Test Helpers

Update `helpers/test_setups.go` to include your new backend:

```go
func SetupDB(t *testing.T, name string) zerokv.Core {
    tmp := t.TempDir()
    var db zerokv.Core
    var err error
    
    switch name {
    case "badgerdb":
        db, err = badgerdb.NewBadgerDB(badgerdb.Config{Dir: tmp})
    case "pebbledb":
        db, err = pebbledb.NewPebbleDB(pebbledb.Config{Dir: tmp})
    case "newdb":
        db, err = newdb.NewNewDB(newdb.Config{Dir: tmp})
    default:
        t.Fatalf("Unknown database: %s", name)
    }
    
    if err != nil || db == nil {
        t.Fatalf("Failed to create %s: %v", name, err)
    }
    return db
}
```

---

## Testing Requirements

### Running Tests

```bash
# Run all tests with coverage
go test ./... -cover

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### What Must Be Tested

Every implementation must pass:

1. **CRUD Tests** (`tests/crud_test.go`)
   - Put/Get/Delete operations
   - Non-existent key retrieval
   - Key overwriting

2. **Iterator Tests** (`tests/iterator_test.go`)
   - Iteration with prefix
   - Key existence checks
   - Iterator release

3. **Batch Tests** (implementation-specific)
   - Batch Put operations
   - Batch Commit operations
   - Error handling for reused batches

### Test Coverage Goals

- Minimum: 80% code coverage
- All error paths should be tested
- Edge cases must be covered (empty data, context cancellation, etc.)

---

## Using ZeroKV

### Basic Usage

```go
package main

import (
    "context"
    "github.com/rawbytedev/zerokv/badgerdb"
)

func main() {
    // Initialize database
    db, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/mydb"})
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Put data
    err = db.Put(ctx, []byte("key"), []byte("value"))
    if err != nil {
        panic(err)
    }
    
    // Get data
    value, err := db.Get(ctx, []byte("key"))
    if err != nil {
        panic(err)
    }
    
    // Delete data
    err = db.Delete(ctx, []byte("key"))
    if err != nil {
        panic(err)
    }
}
```

### Batch Operations

```go
ctx := context.Background()
batch := db.Batch()

// Add operations to batch
batch.Put([]byte("key1"), []byte("value1"))
batch.Put([]byte("key2"), []byte("value2"))
batch.Delete([]byte("key3"))

// Commit batch
err := batch.Commit(ctx)
if err != nil {
    panic(err)
}
```

### Iteration with Prefix

```go
ctx := context.Background()
iterator := db.Scan([]byte("prefix_"))

defer iterator.Release()

for iterator.Next() {
    key := iterator.Key()
    value := iterator.Value()
    // Process key-value pair
}

if iterator.Error() != nil {
    panic(iterator.Error())
}
```

### Switching Databases at Runtime

```go
// Using BadgerDB
db, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/data"})
defer db.Close()

// Switch to PebbleDB without changing code
// db, _ := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp/data"})

db.Put(ctx, key, value)
result, _ := db.Get(ctx, key)
```

---

## Pull Request Checklist

Before submitting a pull request, ensure:

### Code Quality

- [ ] Code follows the style guidelines above
- [ ] All exported functions have documentation comments
- [ ] No unused variables or imports (`go vet ./...`)
- [ ] Code is properly formatted (`go fmt ./...`)
- [ ] Race conditions are checked (`go test -race ./...`)

### Testing

- [ ] All tests pass: `go test ./...`
- [ ] New tests are added for new features
- [ ] Test coverage is at least 80%
- [ ] Implementation passes shared integration tests in `tests/`

### Implementation Completeness

- [ ] All `zerokv.Core` interface methods are implemented
- [ ] All `zerokv.Batch` interface methods are implemented
- [ ] All `zerokv.Iterator` interface methods are implemented
- [ ] Context handling is consistent across all methods
- [ ] Iterator `Error()` method handles empty error slices
- [ ] Resource cleanup is properly documented

### Documentation

- [ ] README updated if adding new features
- [ ] Code comments explain non-obvious logic
- [ ] Examples provided for new functionality
- [ ] CONTRIBUTING.md updated if process changes

### Commits

- [ ] Commits are logically organized
- [ ] Commit messages are descriptive
- [ ] No merge commits (rebase before PR)

### For New Backend Implementations

- [ ] Package structure follows the pattern in existing backends
- [ ] Configuration struct defined in `options.go`
- [ ] All CRUD and batch tests pass
- [ ] Iterator tests pass
- [ ] Implementation-specific tests added
- [ ] Helper `SetupDB()` updated to include new backend

---

## Getting Help

- Review existing implementations in `badgerdb/` and `pebbledb/`
- Check the `examples/` directory for usage patterns
- Read the interface definition in `interface.go`
- Run tests with verbose output: `go test ./... -v`

---

## License

By contributing to ZeroKV, you agree that your contributions will be licensed under the Apache License 2.0.
