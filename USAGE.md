# Usage Guide for ZeroKV

This guide shows you how to use ZeroKV in your Go applications.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [CRUD Operations](#crud-operations)
- [Batch Operations](#batch-operations)
- [Iteration](#iteration)
- [Switching Databases](#switching-databases)
- [Error Handling](#error-handling)
- [Complete Examples](#complete-examples)

## Installation

```bash
go get github.com/rawbytedev/zerokv
```

## Basic Usage

ZeroKV provides a minimal abstraction over key-value stores. Here's the simplest example:

```go
package main

import (
    "context"
    "github.com/rawbytedev/zerokv/implementations/badgerdb"
)

func main() {
    // Initialize database
    db, err := badgerdb.NewBadgerDB(badgerdb.Config{
        Dir: "/tmp/myapp_db",
    })
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // Create a context
    ctx := context.Background()
    
    // Use the database
    err = db.Put(ctx, []byte("name"), []byte("Alice"))
    if err != nil {
        panic(err)
    }
}
```

## CRUD Operations

ZeroKV supports the four basic CRUD operations: Create, Read, Update, and Delete.

### Put (Create/Update)

```go
ctx := context.Background()

// Put a key-value pair
err := db.Put(ctx, []byte("user:1"), []byte("Alice"))
if err != nil {
    log.Fatal(err)
}

// Updating an existing key
err = db.Put(ctx, []byte("user:1"), []byte("Bob"))
if err != nil {
    log.Fatal(err)
}
```

### Get (Read)

```go
ctx := context.Background()

// Get a value
value, err := db.Get(ctx, []byte("user:1"))
if err != nil {
    // Error could mean key not found or other I/O error
    log.Fatal(err)
}

fmt.Printf("Value: %s\n", string(value))
```

### Delete

```go
ctx := context.Background()

// Delete a key
err := db.Delete(ctx, []byte("user:1"))
if err != nil {
    log.Fatal(err)
}
```

### Close

Always close the database when done:

```go
err := db.Close()
if err != nil {
    log.Fatal(err)
}
```

## Batch Operations

Batch operations allow you to perform multiple operations atomically:

```go
ctx := context.Background()

// Create a batch
batch := db.Batch()

// Add operations to the batch
batch.Put([]byte("user:1"), []byte("Alice"))
batch.Put([]byte("user:2"), []byte("Bob"))
batch.Put([]byte("user:3"), []byte("Charlie"))
batch.Delete([]byte("user:4"))

// Commit the batch
err := batch.Commit(ctx)
if err != nil {
    log.Fatal(err)
}

// All operations are now persisted
```

### Important: Batch Behavior After Commit

**BadgerDB**: Returns an error if you try to use a batch after commit

```go
batch := db.Batch()
batch.Put([]byte("key"), []byte("value"))
batch.Commit(ctx)

// This will return an error
err := batch.Put([]byte("key2"), []byte("value2"))
// "This transaction has been discarded. Create a new one"
```

**PebbleDB**: Panics if you try to use a batch after commit

```go
batch := db.Batch()
batch.Put([]byte("key"), []byte("value"))
batch.Commit(ctx)

// This will panic
batch.Put([]byte("key2"), []byte("value2")) // panic!
```

See [ERROR_HANDLING.md](ERROR_HANDLING.md) for more details on implementation-specific error behavior.

## Iteration

ZeroKV provides an iterator interface for scanning keys with a prefix:

```go
ctx := context.Background()

// Populate some data
batch := db.Batch()
batch.Put([]byte("user:1"), []byte("Alice"))
batch.Put([]byte("user:2"), []byte("Bob"))
batch.Put([]byte("user:3"), []byte("Charlie"))
batch.Put([]byte("post:1"), []byte("Hello World"))
batch.Commit(ctx)

// Scan all keys with prefix "user:"
iterator := db.Scan([]byte("user:"))
defer iterator.Release()

for iterator.Next() {
    key := iterator.Key()
    value := iterator.Value()
    fmt.Printf("%s = %s\n", string(key), string(value))
}

// Check for errors during iteration
if iterator.Error() != nil {
    log.Fatal(iterator.Error())
}
```

### Iterator Output

```txt
user:1 = Alice
user:2 = Bob
user:3 = Charlie
```

## Switching Databases

One of ZeroKV's key benefits is the ability to switch databases without changing your code:

```go
package main

import (
    "context"
    "github.com/rawbytedev/zerokv"
    "github.com/rawbytedev/zerokv/badgerdb"
    "github.com/rawbytedev/zerokv/pebbledb"
)

// Function that works with any zerokv.Core implementation
func storeData(db zerokv.Core, key, value string) error {
    return db.Put(context.Background(), 
        []byte(key), 
        []byte(value))
}

func main() {
    var db zerokv.Core
    var err error
    
    // Using BadgerDB
    db, err = badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/data"})
    if err != nil {
        panic(err)
    }
    
    // Store some data
    storeData(db, "name", "Alice")
    
    // Switch to PebbleDB - no other code changes needed!
    db.Close()
    db, err = pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp/data2"})
    if err != nil {
        panic(err)
    }
    
    storeData(db, "name", "Bob")
    db.Close()
}
```

## Using MemDB for Testing

MemDB is an in-memory implementation perfect for testing and development:

```go
package main

import (
    "context"
    "github.com/rawbytedev/zerokv/implementations/memdb"
)

func main() {
    // No configuration needed - data stays in memory
    db, err := memdb.NewMemDB(memdb.Config{})
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Use like any other ZeroKV database
    db.Put(ctx, []byte("test"), []byte("value"))
    value, _ := db.Get(ctx, []byte("test"))
    println(string(value)) // Output: value
}
```

**Note:** MemDB does not support the `Scan` operation for iteration.

## Error Handling

Different database implementations handle errors differently. Always check the [ERROR_HANDLING.md](ERROR_HANDLING.md) documentation for your specific implementation.

### General Error Handling

```go
ctx := context.Background()

// Get might return an error if:
// 1. Key not found
// 2. I/O error
// 3. Context cancelled
value, err := db.Get(ctx, []byte("nonexistent"))
if err != nil {
    log.Printf("Error: %v\n", err)
}
```

### Context Cancellation

ZeroKV respects context cancellation:

```go
// Create a context that expires in 100ms
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

// If the operation takes longer, it will return context error
value, err := db.Get(ctx, []byte("key"))
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Operation timed out")
    } else {
        log.Println("Other error:", err)
    }
}
```

## Complete Examples

### Example 1: Simple User Store

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "github.com/rawbytedev/zerokv/badgerdb"
)

type User struct {
    ID   string
    Name string
    Age  int
}

func main() {
    db, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/users"})
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Store a user
    user := User{ID: "1", Name: "Alice", Age: 30}
    data, _ := json.Marshal(user)
    db.Put(ctx, []byte("user:1"), data)
    
    // Retrieve a user
    value, _ := db.Get(ctx, []byte("user:1"))
    var retrieved User
    json.Unmarshal(value, &retrieved)
    
    log.Printf("User: %+v\n", retrieved)
}
```

### Example 2: Batch Insert with Iteration

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/rawbytedev/zerokv/pebbledb"
)

func main() {
    db, err := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp/cache"})
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Batch insert
    batch := db.Batch()
    for i := 1; i <= 100; i++ {
        key := []byte(fmt.Sprintf("key:%03d", i))
        value := []byte(fmt.Sprintf("value-%d", i))
        batch.Put(key, value)
    }
    batch.Commit(ctx)
    
    // Iterate and count
    iterator := db.Scan([]byte("key:"))
    defer iterator.Release()
    
    count := 0
    for iterator.Next() {
        count++
    }
    
    fmt.Printf("Total keys: %d\n", count)
}
```

### Example 3: Fallback to Backup DB

```go
package main

import (
    "context"
    "log"
    "github.com/rawbytedev/zerokv"
    "github.com/rawbytedev/zerokv/badgerdb"
    "github.com/rawbytedev/zerokv/pebbledb"
)

func getValue(primary, backup zerokv.Core, key []byte) ([]byte, error) {
    // Try primary database first
    value, err := primary.Get(context.Background(), key)
    if err == nil {
        return value, nil
    }
    
    log.Printf("Primary DB failed: %v, trying backup\n", err)
    
    // Fall back to backup database
    return backup.Get(context.Background(), key)
}

func main() {
    primary, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/primary"})
    backup, _ := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp/backup"})
    
    defer primary.Close()
    defer backup.Close()
    
    ctx := context.Background()
    
    // Store in both
    primary.Put(ctx, []byte("key"), []byte("value"))
    backup.Put(ctx, []byte("key"), []byte("backup-value"))
    
    // Retrieve with fallback
    value, err := getValue(primary, backup, []byte("key"))
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Value: %s\n", string(value))
}
```

## Best Practices

1. **Always defer Close()**: Ensure database resources are released

   ```go
   defer db.Close()
   ```

2. **Always defer iterator.Release()**: Prevent resource leaks

   ```go
   iterator := db.Scan(prefix)
   defer iterator.Release()
   ```

3. **Check iterator.Error()**: Always check for errors after iteration

   ```go
   if iterator.Error() != nil {
       log.Fatal(iterator.Error())
   }
   ```

4. **Use context timeouts**: Prevent operations from hanging

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

5. **Commit batches**: Uncommitted batches won't persist data

   ```go
   batch := db.Batch()
   // ... add operations ...
   batch.Commit(ctx) // Must call this!
   ```

6. **Serialize complex data**: ZeroKV works with raw bytes

   ```go
   // Good: manually serialize
   data := json.Marshal(myStruct)
   db.Put(ctx, key, data)
   
   // Bad: don't assume string conversion works for all data types
   db.Put(ctx, key, []byte(myStruct)) // Type error!
   ```
