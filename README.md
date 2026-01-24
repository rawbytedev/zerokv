# KVStore - Minimal Key-Value Abstraction for Go

A minimal, zero-overhead key-value store abstraction for Go. Write your data logic once, then choose your database by changing imports.

## Philosophy

- Zero-overhead abstraction
- Raw []byte values (you control serialization): You choose serialization (JSON, protobuf, msgpack, custom)
- Context-aware API: Consistent with Go conventions, even for embedded stores
- Minimal dependencies

## Quick Start

```go
package main

import (
    "context"
    "github.com/rawbytedev/zerokv/badgerdb"
)

func main() {
    db, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/temp"})
    defer db.Close()
    
    db.Put(context.Background(), []byte("hello"), []byte("world"))
}
```

## Implementations

- Badger - High-performance embedded KV
- Pebble - RocksDB-inspired embedded store

## Creating Your Own

See CONTRIBUTING.md for implementation guidelines.
