# API Reference

Complete API documentation for ZeroKV interfaces and types.

## Table of Contents

- [Core Interface](#core-interface)
- [Batch Interface](#batch-interface)
- [Iterator Interface](#iterator-interface)
- [Error Handling](#error-handling)
- [Context Support](#context-support)

## Core Interface

The `Core` interface is the entry point for database operations.

```go
type Core interface {
    Put(ctx context.Context, key []byte, data []byte) error
    Get(ctx context.Context, key []byte) ([]byte, error)
    Delete(ctx context.Context, key []byte) error
    Batch() Batch
    Scan(prefix []byte) Iterator
    Close() error
}
```

### Methods

#### Put

```go
func (c Core) Put(ctx context.Context, key []byte, data []byte) error
```

Inserts or updates a key-value pair in the database.

**Parameters:**

- `ctx` - Context for cancellation and deadlines
- `key` - The key to insert or update (not nil)
- `data` - The value to store (can be nil or empty)

**Returns:**

- `nil` on success
- `error` on failure (invalid key, I/O error, context cancelled, etc.)

**Example:**

```go
err := db.Put(context.Background(), []byte("name"), []byte("Alice"))
if err != nil {
    log.Fatal(err)
}
```

**Behavior:**

- If the key already exists, its value is overwritten
- No error is returned if the key already exists
- Values are stored as-is; serialization is your responsibility
- Empty values (len=0) are valid

#### Get

```go
func (c Core) Get(ctx context.Context, key []byte) ([]byte, error)
```

Retrieves the value for a given key.

**Parameters:**

- `ctx` - Context for cancellation and deadlines
- `key` - The key to retrieve (not nil)

**Returns:**

- `([]byte, nil)` on success
- `(nil, error)` if key not found or on I/O error

**Example:**

```go
value, err := db.Get(context.Background(), []byte("name"))
if err != nil {
    log.Println("Key not found or error:", err)
    return
}
fmt.Printf("Value: %s\n", string(value))
```

**Behavior:**

- Returns a copy of the value; modifying it won't affect stored data
- Key not found returns an error (implementation-specific error type)
- Respects context cancellation
- Do NOT modify the returned slice

#### Delete

```go
func (c Core) Delete(ctx context.Context, key []byte) error
```

Deletes a key-value pair from the database.

**Parameters:**

- `ctx` - Context for cancellation and deadlines
- `key` - The key to delete

**Returns:**

- `nil` on success (even if key didn't exist)
- `error` on I/O error or context cancellation

**Example:**

```go
err := db.Delete(context.Background(), []byte("name"))
if err != nil {
    log.Fatal(err)
}
```

**Behavior:**

- If key doesn't exist, no error is returned
- Operation is atomic
- Respects context cancellation

#### Batch

```go
func (c Core) Batch() Batch
```

Creates a new batch for atomic multi-operation writes.

**Returns:**

- A new `Batch` instance

**Example:**

```go
batch := db.Batch()
batch.Put([]byte("key1"), []byte("value1"))
batch.Put([]byte("key2"), []byte("value2"))
err := batch.Commit(context.Background())
```

**Behavior:**

- Returns a new batch instance each time
- Batches are not reusable after `Commit()`
- Batches are not thread-safe
- See `Batch` interface for details

#### Scan

```go
func (c Core) Scan(prefix []byte) Iterator
```

Returns an iterator for keys with the given prefix.

**Parameters:**

- `prefix` - The key prefix to scan (can be empty for all keys)

**Returns:**

- An `Iterator` instance

**Example:**

```go
iter := db.Scan([]byte("user:"))
if iter == nil {
    log.Fatal("Scanning not supported by this backend")
}
defer iter.Release()
for iter.Next() {
    key := iter.Key()
    value := iter.Value()
    log.Printf("%s = %s\n", string(key), string(value))
}
if iter.Error() != nil {
    log.Fatal(iter.Error())
}
```

**Behavior:**

- Prefix matching is lexicographic
- Empty prefix matches all keys
- Must call `Release()` on the returned iterator
- May return `nil` if the backend doesn't support scanning (e.g., MemDB)
- Always check if iterator is `nil` before use
- See `Iterator` interface for details

#### Close

```go
func (c Core) Close() error
```

Closes the database and releases all resources.

**Returns:**

- `nil` on successful close
- `error` if close operation fails

**Example:**

```go
err := db.Close()
if err != nil {
    log.Fatal(err)
}
```

**Behavior:**

- Flushes pending writes
- Closes file handles
- Releases memory
- Should be called in a defer statement
- No operations should be performed after Close()

---

## Batch Interface

The `Batch` interface groups multiple operations for atomic writes.

```go
type Batch interface {
    Put(key []byte, data []byte) error
    Delete(key []byte) error
    Commit(ctx context.Context) error
}
```

### Batch Methods

#### Put (Batch)

```go
func (b Batch) Put(key []byte, data []byte) error
```

Adds a key-value pair to the batch.

**Parameters:**

- `key` - The key
- `data` - The value

**Returns:**

- `nil` on success
- `error` if batch is already committed or other error

**Example:**

```go
batch := db.Batch()
err := batch.Put([]byte("key1"), []byte("value1"))
if err != nil {
    log.Fatal(err)
}
```

**Behavior:**

- Adds operation to the batch queue
- Does NOT write to database yet
- Cannot be used after `Commit()`
- Later operations with same key override earlier ones

#### Delete (Batch)

```go
func (b Batch) Delete(key []byte) error
```

Adds a delete operation to the batch.

**Parameters:**

- `key` - The key to delete

**Returns:**

- `nil` on success
- `error` if batch is already committed or other error

**Example:**

```go
batch := db.Batch()
batch.Delete([]byte("old_key"))
batch.Commit(context.Background())
```

**Behavior:**

- Adds delete operation to batch queue
- Does NOT delete from database yet
- Cannot be used after `Commit()`
- Deleting non-existent keys is allowed

#### Commit

```go
func (b Batch) Commit(ctx context.Context) error
```

Writes all batched operations to the database atomically.

**Parameters:**

- `ctx` - Context for cancellation and deadlines

**Returns:**

- `nil` on successful commit
- `error` if commit fails or context is cancelled

**Example:**

```go
batch := db.Batch()
batch.Put([]byte("key1"), []byte("value1"))
batch.Put([]byte("key2"), []byte("value2"))
err := batch.Commit(context.Background())
if err != nil {
    log.Fatal(err)
}
```

**Behavior:**

- Writes all operations atomically
- Either all operations succeed or none
- Cannot be called twice on the same batch
- Batch cannot be reused after `Commit()`
- Respects context cancellation

**Important Notes on Batch Reuse:**

**BadgerDB:** Attempting to use a batch after `Commit()` returns an error

```go
batch.Put(key, value)
batch.Commit(ctx)
err := batch.Put(key2, value2) // Returns error
```

**PebbleDB:** Attempting to use a batch after `Commit()` panics

```go
batch.Put(key, value)
batch.Commit(ctx)
batch.Put(key2, value2) // Panics!
```

**Always create a new batch if you need more operations:**

```go
batch1 := db.Batch()
batch1.Put(key1, value1)
batch1.Commit(ctx)

batch2 := db.Batch() // New batch!
batch2.Put(key2, value2)
batch2.Commit(ctx)
```

---

## Iterator Interface

The `Iterator` interface provides sequential access to key-value pairs.

```go
type Iterator interface {
    Next() bool
    Key() []byte
    Value() []byte
    Release()
    Error() error
}
```

### iterator Methods

#### Next

```go
func (it Iterator) Next() bool
```

Advances the iterator to the next key-value pair.

**Returns:**

- `true` if there is a next item
- `false` if iteration is complete

**Example:**

```go
iter := db.Scan([]byte("user:"))
defer iter.Release()
for iter.Next() {
    key := iter.Key()
    value := iter.Value()
    log.Printf("%s = %s\n", string(key), string(value))
}
```

**Behavior:**

- On first call, seeks to the first matching key
- On subsequent calls, advances to the next key
- Returns `false` when no more items
- After `false`, `Key()` and `Value()` return nil

#### Key

```go
func (it Iterator) Key() []byte
```

Returns the key of the current item.

**Returns:**

- `[]byte` containing the key
- `nil` if iterator is not on a valid item

**Example:**

```go
for iter.Next() {
    key := iter.Key()
    if key != nil {
        log.Printf("Key: %s\n", string(key))
    }
}
```

**Behavior:**

- Returns a copy of the key
- Safe to use after iteration moves to next item
- Returns nil if `Next()` returned false
- Returns nil before first `Next()` call

#### Value

```go
func (it Iterator) Value() []byte
```

Returns the value of the current item.

**Returns:**

- `[]byte` containing the value
- `nil` if iterator is not on a valid item or error occurred

**Example:**

```go
for iter.Next() {
    value := iter.Value()
    if value != nil {
        log.Printf("Value: %s\n", string(value))
    }
}
```

**Behavior:**

- Returns a copy of the value
- Safe to use after iteration moves to next item
- Returns nil if `Next()` returned false
- Returns nil before first `Next()` call
- May return nil if error occurs during value retrieval

#### Release

```go
func (it Iterator) Release()
```

Closes the iterator and releases all resources.

**Example:**

```go
iter := db.Scan(prefix)
defer iter.Release()
for iter.Next() {
    // Use iterator
}
```

**Behavior:**

- Closes underlying database iterator
- Frees associated resources
- Should always be called (use defer)
- Safe to call multiple times
- Safe to call even if iteration incomplete

**IMPORTANT:** Always defer Release():

```go
// Good
iter := db.Scan(prefix)
defer iter.Release()

// Bad - resource leak if error occurs
iter := db.Scan(prefix)
// ... code ...
iter.Release() // May not be called if error occurs earlier
```

#### Error

```go
func (it Iterator) Error() error
```

Returns the last error that occurred during iteration.

**Returns:**

- `nil` if no error has occurred
- `error` containing the last error

**Example:**

```go
iter := db.Scan(prefix)
defer iter.Release()
for iter.Next() {
    key := iter.Key()
    value := iter.Value()
}

// Check for errors
if iter.Error() != nil {
    log.Fatal("Iteration error:", iter.Error())
}
```

**Behavior:**

- Returns `nil` if iteration was successful
- Returns the most recent error if one occurred
- Safe to call multiple times
- Safe to call even if iteration was incomplete
- Never panics (safe error handling)

---

## TTL Implementation

### TTLProvider Interface

The `TTLProvider` provide 
```go
type TTLProvider interface {
	TTLPut(ctx context.Context, key, data []byte, ttl time.Duration) error
}
```

**TTLPut** : Insert an entry with specified ttl(duration to live)

```go
type BatchWithTTL interface {
	zerokv.Batch
	PutWithTTL(key []byte, data []byte, ttl time.Duration) error
}
```

**PutWithTTL** : Insert an entry with specified ttl in a batch - it requires a commit later

## Error Handling

### Return Values

All methods follow Go conventions:

- `(value, nil)` on success
- `(zero-value, error)` on failure

### Error Types

Errors are implementation-specific. Check [ERROR_HANDLING.md](ERROR_HANDLING.md) for details on how each implementation handles errors.

Common errors:

- Key not found (from `Get()`)
- I/O errors (from underlying database)
- Context cancelled errors
- Invalid parameters

### Error Checking

```go
// Always check errors
value, err := db.Get(ctx, key)
if err != nil {
    // Handle error
}

// For batch operations
batch := db.Batch()
if err := batch.Put(key, value); err != nil {
    // Handle error
}
if err := batch.Commit(ctx); err != nil {
    // Handle error
}
```

---

## Context Support

All operations accept a `context.Context` parameter for:

### Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel operation
cancel()

// Operation will return context.Cancelled error
value, err := db.Get(ctx, key)
if errors.Is(err, context.Cancelled) {
    log.Println("Operation was cancelled")
}
```

### Timeouts

```go
// 5-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := db.Get(ctx, key)
if errors.Is(err, context.DeadlineExceeded) {
    log.Println("Operation timed out")
}
```

### Deadlines

```go
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

value, err := db.Get(ctx, key)
```

**All operations check context at the start:**

- If context is already cancelled, returns immediately
- Respects deadline/timeout of context
- Returns appropriate context error

---

## Example: Complete Usage

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/rawbytedev/zerokv/badgerdb"
)

func main() {
    // Initialize
    db, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/tmp/mydb"})
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Put
    db.Put(ctx, []byte("user:1"), []byte("Alice"))
    db.Put(ctx, []byte("user:2"), []byte("Bob"))
    
    // Get
    value, _ := db.Get(ctx, []byte("user:1"))
    log.Printf("User 1: %s\n", string(value))
    
    // Batch
    batch := db.Batch()
    batch.Put([]byte("user:3"), []byte("Charlie"))
    batch.Put([]byte("user:4"), []byte("David"))
    batch.Commit(ctx)
    
    // Scan
    iter := db.Scan([]byte("user:"))
    defer iter.Release()
    for iter.Next() {
        log.Printf("%s = %s\n", string(iter.Key()), string(iter.Value()))
    }
    
    // Delete
    db.Delete(ctx, []byte("user:4"))
    
    // Close
    db.Close()
}
```

**Output:**

```txt
User 1: Alice
user:1 = Alice
user:2 = Bob
user:3 = Charlie
user:4 = David
```
