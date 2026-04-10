# ZeroKV - Minimal Key-Value Store Abstraction for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/rawbytedev/zerokv)](https://goreportcard.com/report/github.com/rawbytedev/zerokv)
![Test and Benchmark](https://github.com/rawbytedev/zerokv/actions/workflows/tests.yml/badge.svg)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rawbytedev/zerokv)](https://github.com/rawbytedev/zerokv)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/rawbytedev/zerokv.svg)](https://pkg.go.dev/github.com/rawbytedev/zerokv)
[![GitHub last commit](https://img.shields.io/github/last-commit/rawbytedev/zerokv)](https://github.com/rawbytedev/zerokv)
[![GitHub Release](https://img.shields.io/github/v/release/rawbytedev/zerokv)](https://github.com/rawbytedev/zerokv/releases/latest)
[![GitHub issues](https://img.shields.io/github/issues/rawbytedev/zerokv)](https://github.com/rawbytedev/zerokv/issues)

**ZeroKV** is a minimal, zero-overhead key-value store abstraction for Go. Write your data logic once, then **switch databases by changing a single import**. Perfect for applications that need flexibility in database choice.

## Key Features

- **Minimal Abstraction** - Small, focused interface (Core, Batch, Iterator)
- **Zero Overhead** - No reflection, no wrappers, just direct calls
- **Raw Bytes** - Full control over serialization (JSON, Protobuf, msgpack, etc.)
- **Context Aware** - Full support for cancellation, timeouts, and deadlines
- **Pluggable Backends** - Switch databases without changing application code
- **Production Ready** - Well-tested with multiple implementations
- **Comprehensive Testing** - 80%+ test coverage with shared integration tests

## Philosophy

ZeroKV is built on three core principles:

1. **Minimal API** - Only essential operations: Put, Get, Delete, Batch, Scan, Close
2. **Zero Wrapper Overhead** - Direct database calls, no unnecessary abstractions
3. **You Control Serialization** - Work with raw bytes, serialize however you want
4. **Context Support** - Proper Go idioms with context.Context
5. **Explicit Over Implicit** - Clear behavior, no hidden complexity

## Quick Start

### Installation

```bash
go get github.com/rawbytedev/zerokv
```

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

    // Put - Insert or update
    db.Put(ctx, []byte("hello"), []byte("world"))

    // Get - Retrieve
    value, err := db.Get(ctx, []byte("hello"))
    if err == nil {
        println(string(value)) // Output: world
    }

    // Delete - Remove
    db.Delete(ctx, []byte("hello"))
}
```

### Switching Databases

One of ZeroKV's superpowers is database portability:

```go
// Using BadgerDB
db, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/data"})

// Switch to PebbleDB - zero code changes needed
// db, _ := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp/data"})

// All existing code works unchanged
db.Put(ctx, []byte("key"), []byte("value"))
```

## Implementations

### Built-in Backends

| Database | Features | Best For |
| ---------- | ---------- | ---------- |
| **BadgerDB** | High-performance LSM tree | Write-heavy workloads, strong consistency |
| **PebbleDB** | RocksDB-compatible, flexible | Read-heavy workloads, compatibility needs |
| **MemDB** | In-memory, fast | Testing, caching, development (no Scan support) |

### Custom Implementations

Implementing your own backend is straightforward. See [IMPLEMENTATION.md](IMPLEMENTATION.md) for a complete guide.

## Documentation

- **[USAGE.md](USAGE.md)** - Complete usage guide with examples
- **[API.md](API.md)** - Full API reference
- **[ERROR_HANDLING.md](ERROR_HANDLING.md)** - Error behavior across implementations
- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Guide for implementing new backends
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines

## Core Operations

### CRUD Operations

```go
// Put - Insert or update
db.Put(ctx, []byte("name"), []byte("Alice"))

// Get - Retrieve
value, err := db.Get(ctx, []byte("name"))

// Delete - Remove
db.Delete(ctx, []byte("name"))
```

### Batch Operations

```go
batch := db.Batch()
batch.Put([]byte("user:1"), []byte("Alice"))
batch.Put([]byte("user:2"), []byte("Bob"))
batch.Delete([]byte("old_user"))
err := batch.Commit(ctx)
```

### Iteration

```go
iter := db.Scan([]byte("user:"))
defer iter.Release()

for iter.Next() {
    key := iter.Key()
    value := iter.Value()
    // Process key-value pair
}

if iter.Error() != nil {
    panic(iter.Error())
}
```

## Advanced Features

### Context Support

```go
// Timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := db.Get(ctx, []byte("key"))
if errors.Is(err, context.DeadlineExceeded) {
    log.Println("Operation timed out")
}

// Cancellation
ctx, cancel := context.WithCancel(context.Background())
cancel()

err := db.Put(ctx, []byte("key"), []byte("value"))
// Returns context.Cancelled error
```

### Error Handling

Different backends may handle errors differently. Always refer to [ERROR_HANDLING.md](ERROR_HANDLING.md):

```go
// BadgerDB returns error on batch reuse
batch.Put(key, value)
batch.Commit(ctx)
err := batch.Put(key2, value2) // Error

// PebbleDB panics on batch reuse
batch.Put(key, value)
batch.Commit(ctx)
batch.Put(key2, value2) // Panic - create new batch!
```

## Performance

ZeroKV adds minimal overhead:

- No reflection or interface{} conversions
- Direct method calls to underlying database
- Zero-copy where possible
- Efficient memory handling

## Project Structure

```js
zerokv/
├── core/
│   └── interface.go     # Core interfaces
├── internal/            # Internal utilities
├── implementations/
│   ├── badgerdb/        # BadgerDB implementation
│   │   ├── badgerdb.go  # Main code
│   │   ├── badgerdb_test.go # Tests
│   │   └── options.go   # Configuration
│   ├── pebbledb/        # PebbleDB implementation
│   │   ├── pebbledb.go
│   │   ├── pebbledb_test.go
│   └── memdb/           # MemDB implementation
│       ├── mem.go
│       ├── mem_test.go
│       └── options.go
├── serializers/         # Serialization helpers
├── helpers/             # Test utilities
├── testing/             # Shared integration tests
├── examples/            # Usage examples
```

## Testing

Comprehensive test coverage across implementations:

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run with race detector
go test ./... -race

# Run specific implementation
go test ./badgerdb -v
go test ./pebbledb -v
```

## Examples

### Example 1: User Store with JSON

```go
import "encoding/json"

type User struct {
    ID   string
    Name string
    Age  int
}

// Store user
user := User{ID: "1", Name: "Alice", Age: 30}
data, _ := json.Marshal(user)
db.Put(ctx, []byte("user:1"), data)

// Retrieve and deserialize
value, _ := db.Get(ctx, []byte("user:1"))
var retrieved User
json.Unmarshal(value, &retrieved)
```

### Example 2: Batch Insert

```go
batch := db.Batch()
for i := 0; i < 1000; i++ {
    key := []byte(fmt.Sprintf("key:%d", i))
    value := []byte(fmt.Sprintf("value:%d", i))
    batch.Put(key, value)
}
batch.Commit(ctx)
```

### Example 3: Database Failover

```go
func getValue(primary, backup zerokv.Core, key []byte) ([]byte, error) {
    value, err := primary.Get(ctx, key)
    if err == nil {
        return value, nil
    }
    
    log.Println("Primary failed, trying backup")
    return backup.Get(ctx, key)
}
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### To Implement a New Backend

1. Read [IMPLEMENTATION.md](IMPLEMENTATION.md)
2. Create your package (e.g., `mydb/`)
3. Implement `Core`, `Batch`, and `Iterator` interfaces
4. Add tests
5. Submit a pull request

### Requirements

- **Go 1.25.2+**
- Chosen Database

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Support

- Documentation: See [USAGE.md](USAGE.md) and [API.md](API.md)
- Issues: [GitHub Issues](https://github.com/rawbytedev/zerokv/issues)
- Discussions: [GitHub Discussions](https://github.com/rawbytedev/zerokv/discussions)
