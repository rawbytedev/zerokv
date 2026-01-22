package dbmodel

type Store interface {
	// Put inserts or updates a key-value pair in the database.
	Put(key []byte, data []byte) error
	// Get retrieves the value for a given key. Returns an error if not found.
	Get(key []byte) ([]byte, error)
	// Del deletes a key-value pair from the database.
	Delete(key []byte) error
	// Batch Operation creates a new batch operation for the database.
	Batch() Batch
	// Iterate over Database
	Scan(prefix []byte) Iterator
	// Close closes the database and releases all resources.
	Close() error
}

type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Release()
	Error() error
}

type Batch interface {
	// Write commits the batch operations to the database.
	Commits() error
	// Put inserts or updates a key-value pair in the database.
	Put(key []byte, data []byte) error
	// Del deletes a key-value pair from the database.
	Delete(key []byte) error
}

// mainly for testing and debugging
type Operations struct {
	Key   []byte
	Value []byte
	Type  Ops
}

type Ops int

const (
	PutOp Ops = iota
	GetOp
	DeleteOp
)
