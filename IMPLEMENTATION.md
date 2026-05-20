# Implementation Guide

This guide provides detailed instructions for implementing a new storage backend for ZeroKV.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Step-by-Step Implementation](#step-by-step-implementation)
- [Error Handling Strategy](#error-handling-strategy)
- [Testing Your Implementation](#testing-your-implementation)
- [Integration](#integration)
- [Checklist](#checklist)

## Overview

ZeroKV provides a minimal interface (`zerokv.Core`) that you must implement. Your implementation wraps an underlying database library and exposes it through the ZeroKV interface.

The three main components to implement are:

1. **Core** - Main database interface (Put, Get, Delete, Batch, Scan, Close)
2. **Batch** - Batch operation interface (Put, Delete, Commit)
3. **Iterator** - Iteration interface (Next, Key, Value, Release, Error)

With an optional component:

1. TTL(Time-to-live)

**TTLProvider** - interface for assigning TTL to entries (TTLPut)
**BatchWithTTL** - interface for Batch operations on TTL entries (PutWithTTL)

## Prerequisites

- Go 1.25.2 or higher

## Step-by-Step Implementation

### Step 1: Create Package Structure

Create a new directory for your backend:

```bash
mkdir mydb
cd mydb
touch mydb.go mydb_test.go options.go
```

Directory structure:

```txt
mydb/
├── mydb.go       # Main implementation
├── mydb_test.go  # Implementation-specific tests
└── options.go    # Configuration
```

### Step 2: Define Configuration (`options.go`)

Create a configuration struct for your database:

```go
package mydb

import "github.com/your-org/your-db"

// Config holds configuration for MyDB
type Config struct {
    // Required
    Dir string
    
    // Optional: backend-specific configuration
    MyDBConfigs *yourdb.Options
}

// DefaultOptions returns a Config with sensible defaults
func DefaultOptions(dir string) *Config {
    return &Config{
        Dir:         dir,
        MyDBConfigs: nil, // Use backend defaults
    }
}
```

### Step 3: Define Types (`mydb.go` - Part 1)

Define the structs for Core, Batch, and Iterator:

```go
package mydb

import (
    "context"
    "github.com/rawbytedev/zerokv"
    "github.com/your-org/your-db"
)

// myDB implements zerokv.Core interface
type myDB struct {
    db *yourdb.DB
}

// myBatch implements zerokv.Batch interface
type myBatch struct {
    batch *yourdb.WriteBatch
    // Track if batch is closed to prevent reuse
    closed bool
}

// myIterator implements zerokv.Iterator interface
type myIterator struct {
    Iterator *yourdb.Iterator
    started  bool
    valid    bool
    err      []error
}

```

### Step 4: Implement Constructor

```go
// NewMyDB initializes and returns a zerokv.Core instance
func NewMyDB(cfg Config) (zerokv.Core, error) {
    opts := &yourdb.Options{}
    if cfg.MyDBConfigs != nil {
        opts = cfg.MyDBConfigs
    }
    
    db, err := yourdb.Open(cfg.Dir, opts)
    if err != nil {
        return nil, err
    }
    
    return &myDB{db: db}, nil
}
```

### Step 5: Implement Core Interface Methods

#### Put

```go
// Put inserts or updates a key-value pair in the database
func (m *myDB) Put(ctx context.Context, key []byte, data []byte) error {
    // Always check context first
    if err := ctx.Err(); err != nil {
        return err
    }
    
    // Call underlying database
    return m.db.Set(key, data)
}
```

#### Get

```go
// Get retrieves the value for a given key
// Returns an error if the key is not found
func (m *myDB) Get(ctx context.Context, key []byte) ([]byte, error) {
    // Always check context first
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    
    // Call underlying database
    value, err := m.db.Get(key)
    if err != nil {
        return nil, err
    }
    
    // If your database returns a reference to internal memory,
    // make a copy to prevent data corruption
    result := make([]byte, len(value))
    copy(result, value)
    return result, nil
}
```

#### Delete

```go
// Delete removes a key-value pair from the database
func (m *myDB) Delete(ctx context.Context, key []byte) error {
    // Always check context first
    if err := ctx.Err(); err != nil {
        return err
    }
    
    return m.db.Delete(key)
}
```

#### Batch

```go
// Batch returns a new batch for atomic operations
func (m *myDB) Batch() zerokv.Batch {
    return &myBatch{
        batch:  m.db.NewWriteBatch(),
        closed: false,
    }
}
```

#### Scan

```go
// Scan returns an iterator for keys with the given prefix
func (m *myDB) Scan(prefix []byte) zerokv.Iterator {
    iter := m.db.NewIterator(yourdb.IterOptions{Prefix: prefix})
    return &myIterator{
        Iterator: iter,
        started:  false,
        valid:    false,
        err:      []error{},
    }
}
```

#### Close

```go
// Close closes the database and releases all resources
func (m *myDB) Close() error {
    if m.db != nil {
        return m.db.Close()
    }
    return nil
}
```

### Step 6: Implement Batch Interface Methods

```go
// Put adds a key-value pair to the batch
func (b *myBatch) Put(key []byte, data []byte) error {
    if b.closed {
        return fmt.Errorf("batch already committed, create a new one")
    }
    return b.batch.Set(key, data)
}

// Delete removes a key from the batch
func (b *myBatch) Delete(key []byte) error {
    if b.closed {
        return fmt.Errorf("batch already committed, create a new one")
    }
    return b.batch.Delete(key)
}

// Commit writes all batch operations to the database atomically
func (b *myBatch) Commit(ctx context.Context) error {
    if b.closed {
        return fmt.Errorf("batch already committed")
    }
    
    if err := ctx.Err(); err != nil {
        return err
    }
    
    b.closed = true
    return b.batch.Write()
}
```

### Step 7: Implement Iterator Interface Methods

```go
// Next advances the iterator to the next key
func (it *myIterator) Next() bool {
    if !it.started {
        // First call - seek to beginning
        it.valid = it.Iterator.First()
        it.started = true
    } else {
        // Subsequent calls - advance to next
        it.valid = it.Iterator.Next()
    }
    return it.valid
}

// Key returns the current key
func (it *myIterator) Key() []byte {
    if !it.valid {
        return nil
    }
    
    // Make a copy if your DB returns references
    key := it.Iterator.Key()
    result := make([]byte, len(key))
    copy(result, key)
    return result
}

// Value returns the current value
func (it *myIterator) Value() []byte {
    if !it.valid {
        return nil
    }
    
    value, err := it.Iterator.Value()
    if err != nil {
        it.err = append(it.err, err)
        return nil
    }
    
    // Make a copy if your DB returns references
    result := make([]byte, len(value))
    copy(result, value)
    return result
}

// Release closes the iterator and frees resources
func (it *myIterator) Release() {
    if it.Iterator != nil {
        it.Iterator.Close()
    }
    it.valid = false
}

// Error returns the last error that occurred during iteration
// IMPORTANT: Handle empty slice to avoid panics
func (it *myIterator) Error() error {
    if len(it.err) == 0 {
        return nil
    }
    return it.err[len(it.err)-1]
}
```

### Step 8: Implement TTL(Time To Live) (optional)


```go
// single insertion
func (b *myDB) TTLPut(ctx context.Context, key []byte, value []byte, duration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
    // in this case that's how to set TTL to entries in badger
    // make sure to refer to you're specific DB for correct usage
	return b.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value).WithTTL(duration)
		return txn.SetEntry(e)
	})
}
```

```go
// Batch insertion with ttl
func (b *badgerBatch) PutWithTTL(key []byte, value []byte, ttl time.Duration) error {
	// in this case that's how to set TTL to entries in badger
    // make sure to refer to you're specific DB for correct usage
    e := badger.NewEntry(key, value).WithTTL(ttl)
	return b.batch.SetEntry(e)
}
```

## Error Handling Strategy

### Key Principles

1. **Never panic** - Always return errors from public methods
2. **Check context** - All operations must check `ctx.Err()` first
3. **Document behavior** - Clearly document any backend-specific behavior

### Panic Recovery Pattern (if needed)

If your underlying database panics, catch and convert to error:

```go
func (b *myBatch) Put(key []byte, data []byte) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("underlying database panic: %v", r)
        }
    }()
    
    if b.closed {
        return fmt.Errorf("batch already committed")
    }
    
    return b.batch.Set(key, data)
}
```

## Testing Your Implementation

### Step 1: Add Implementation-Specific Tests

In `mydb_test.go`:

```go
package mydb_test

import (
    "testing"
    "github.com/rawbytedev/zerokv/helpers"
    "github.com/stretchr/testify/require"
    "github.com/rawbytedev/zerokv/mydb"
)

// TestMyDBBatchOperations tests batch Put and Get operations
func TestMyDBBatchOperations(t *testing.T) {
    db := helpers.SetupDB(t, "mydb")
    defer db.Close()
    
    batch := db.Batch()
    keys := make([][]byte, 5)
    values := make([][]byte, 5)
    
    for i := 0; i < 5; i++ {
        keys[i] = helpers.RandomBytes(16)
        values[i] = helpers.RandomBytes(32)
        err := batch.Put(keys[i], values[i])
        require.NoError(t, err)
    }
    
    err := batch.Commit(t.Context())
    require.NoError(t, err)
    
    // Verify all values were written
    for i := 0; i < 5; i++ {
        retrieved, err := db.Get(t.Context(), keys[i])
        require.NoError(t, err)
        require.Equal(t, values[i], retrieved)
    }
}
```

### Step 2: Register in Test Helpers

Update `helpers/test_setups.go`:

```go
package helpers

import (
    // ... existing imports ...
    "github.com/rawbytedev/zerokv/mydb"
)

func SetupDB(t *testing.T, name string) zerokv.Core {
    tmp := t.TempDir()
    var db zerokv.Core
    var err error
    
    switch name {
    case "badgerdb":
        db, err = badgerdb.NewBadgerDB(badgerdb.Config{Dir: tmp})
    case "pebbledb":
        db, err = pebbledb.NewPebbleDB(pebbledb.Config{Dir: tmp})
    case "mydb":
        db, err = mydb.NewMyDB(mydb.Config{Dir: tmp})
    default:
        t.Fatalf("Unknown database: %s", name)
    }
    
    if err != nil || db == nil {
        t.Fatalf("Failed to create %s: %v", name, err)
    }
    return db
}
```

### Step 3: Run All Tests

```bash
# Run shared tests (will now include mydb)
go test ./tests -v

# Run implementation-specific tests
go test ./mydb -v

# Run all tests
go test ./... -v

# Run with race detector
go test ./... -race

# Check coverage
go test ./... -cover
```

## Integration

### Update go.mod

If your implementation has external dependencies, they'll be added to go.mod automatically:

```bash
go mod download
go mod tidy
```

### Create Examples (optional)

Create an example in `examples/mydb_usage.go`:

```go
package main

import (
    "context"
    "github.com/rawbytedev/zerokv/mydb"
)

func mydb_main() {
    db, _ := mydb.NewMyDB(mydb.Config{Dir: "/tmp/mydb"})
    defer db.Close()
    
    ctx := context.Background()
    db.Put(ctx, []byte("hello"), []byte("world"))
    value, _ := db.Get(ctx, []byte("hello"))
    
    println(string(value)) // Output: world
}
```

### Update README.md

Add your implementation to the list:

```markdown
## Implementations

- Badger - High-performance embedded KV
- Pebble - RocksDB-inspired embedded store
- MyDB - Your database description
```

---

## Checklist

Before submitting your implementation, ensure:

### Code Quality

- All methods have documentation comments
- No unused imports or variables (`go vet ./...`)
- Code is formatted (`go fmt ./...`)
- No panics in public methods (Optional)
- All interface methods are implemented

Note: Ensuring Implementations are consistent `return error or turn panics into errors instead of panicking` helps greatly when switching between databases using zerokv but we don't enforce anything it's up to you to decide  

### Core Functionality

- `Put()` works correctly
- `Get()` works and handles missing keys
- `Delete()` works correctly
- `Close()` release resources
- `Batch()` create working batchs

### Batch Operations

- `Batch.Put()` adds to batch
- `Batch.Delete()` removes from batch
- `Batch.Commit()` persists all operations
- Using batch after commit returns error (or panics)

### Iterator Operations

- `Next()` advances iterator correctly
- `Key()` and `Value()` return correct data
- `Release()` closes resources
- `Error()` never panics (returns nil if no error)

### TTL (Time to Live) Operations

- `TTLPut` works if available
- `PutWithTTL` works if available

### Context Handling

- All operations check `ctx.Err()` at the start
- Context cancellation is respected
- Context deadline is respected

### Testing

- Implementation passes shared tests (`tests/`)
- Implementation-specific tests added
- All tests pass: `go test ./...`
- Race detector passes: `go test -race ./...`

### Documentation

- Configuration struct documented
- Error behavior documented
- Example usage provided
- README updated
- Implementation registered in test helpers

### Error Handling

- Error messages are clear and helpful
