package zerokv

import "context"

type Core interface {
	// Put inserts or updates a key-value pair in the database.
	Put(ctx context.Context, key []byte, data []byte) error
	// Get retrieves the value for a given key.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Del deletes a key-value pair from the database.
	Delete(ctx context.Context, key []byte) error
	// Batch Operation creates a new batch operation for the database.
	/*
		Must be used carefully calling Batch creates a new write batch that needs to be committed separately or else it may lead to uncommitted data and data loss.
	*/
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
	/*
		It is crucial to call Commits to ensure that all batched operations are saved to the database.
		and no new insertions/updates/deletions will be saved until Commits is called.
		Inserting to already committed batch is forbidden and will lead to errors.
	*/
	Commit(ctx context.Context) error
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
