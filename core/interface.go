package zerokv

import "context"

// Core defines the main interface for a key-value database
type Core interface {
	// Put inserts or updates a key-value pair in the database
	Put(ctx context.Context, key []byte, data []byte) error
	// Get retrieves the value for a given key
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Delete removes a key-value pair from the database
	Delete(ctx context.Context, key []byte) error
	// Batch creates a new write batch that needs to be committed separately
	Batch() Batch
	// Scan returns an iterator to traverse key-value pairs with the specified prefix
	Scan(prefix []byte) Iterator
	// Close closes the database connection
	Close() error
}

// Iterator defines methods for iterating over key-value pairs in the database
type Iterator interface {
	Next() bool    // advances the iterator to the next key-value pair
	Key() []byte   // returns the current key
	Value() []byte // returns the current value
	Release()      // releases the iterator resources
	Error() error  // returns any error encountered during iteration
}

// Batch defines methods for batching multiple write operations together
type Batch interface {
	// Flush commits all batched operations to the database
	Commit(ctx context.Context) error
	// Put inserts or updates a key-value pair in the database
	Put(key []byte, data []byte) error
	// Delete deletes a key-value pair from the database
	Delete(key []byte) error
}
