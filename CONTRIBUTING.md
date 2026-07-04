# Contributing to ZeroKV

Thank you for your interest in contributing to ZeroKV! This guide provides everything you need to contribute code, implement new backends, and maintain code quality.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Adding New Backend Implementations](#implementing-a-new-storage-backend)
- [Testing Requirements](#testing-requirements)

---

## Getting Started

### Prerequisites

- Go 1.25.2 or higher

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

### First Steps

1. Read [USAGE.md](USAGE.md) to understand how ZeroKV works
2. Review [API.md](API.md) for the complete interface specification
3. Check [ERROR_HANDLING.md](ERROR_HANDLING.md) for error behavior documentation
4. If implementing a backend, read [IMPLEMENTATION.md](IMPLEMENTATION.md)

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

### Building the Project

```bash
# Build all packages
go build ./...

# Run code quality checks
go vet ./...

# Format after making changes
gofmt -w 
```

---

## Code Style Guidelines

### Naming Conventions

- **Interfaces**: Use descriptive names (e.g., `Core`, `Iterator`, `Batch`)
- **Structs**: Use camelCase, no abbreviated names (e.g., `BadgerDB`, not `bdb`)
- **Methods**: Descriptive action verbs (e.g., `Put()`, `Get()`, `Delete()`)
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
func (b *BadgerDB) Put(ctx context.Context, key []byte, data []byte) error {
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
func (b *BadgerDB) Get(ctx context.Context, key []byte) ([]byte, error) {
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

func (n *newDB) Scan(prefix []byte, opts ...ScanOption) zerokv.Iterator {
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
# Run all tests
go test ./...

# Run all tests with coverage
go test ./... -cover

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with race detector
go test ./... -race

# Run specific package tests
go test ./tests -v
go test ./badgerdb -v
go test ./pebbledb -v
```

### What Must Be Tested

Every implementation must pass:

1. **CRUD Tests** (`tests/crud_test.go`)
   - Put/Get/Delete operations
   - Non-existent key retrieval
   - Key overwriting
   - Close operation

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

### Error Handling in Tests

See [ERROR_HANDLING.md](ERROR_HANDLING.md) for how different implementations handle errors and how to test them properly.

---

## Getting Help

- Review existing implementations in `badgerdb/` and `pebbledb/`
- Check the `examples/` directory for usage patterns
- Read [USAGE.md](USAGE.md) for usage examples
- Read [API.md](API.md) for API documentation
- Read [ERROR_HANDLING.md](ERROR_HANDLING.md) for error handling details
- Read [IMPLEMENTATION.md](IMPLEMENTATION.md) for backend implementation guide
- Run tests with verbose output: `go test ./... -v`

---

## License

By contributing to ZeroKV, you agree that your contributions will be licensed under the Apache License 2.0.
