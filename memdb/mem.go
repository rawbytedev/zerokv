package memdb

import (
	"context"

	db "github.com/AmirSoleimani/MemoryDB/memdb"
	"github.com/rawbytedev/zerokv"
)

type MemDB struct {
	db *db.MemDB
}

type MemBatch struct {
	batch db.Batch
}

// NewMemDB initializes and returns a zerokv.Core instance at the specified path(MemDB).
func NewMemDataBase(cfg Config) (zerokv.Core, error) {
	db := db.NewMemDB()
	return &MemDB{db: db}, nil
}

// --- Basic CRUD operations ---

// Put inserts or updates a key-value pair in the database.
func (m *MemDB) Put(ctx context.Context, key []byte, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.db.Put(key, data)
}

// Get retrieves the value for a given key. Returns an error if not found.
func (m *MemDB) Get(ctx context.Context, key []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	val, err := m.db.Get(key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Del deletes a key-value pair from the database.
func (p *MemDB) Delete(ctx context.Context, key []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.db.Delete(key)
}

// Close closes the database and releases all resources.
func (m *MemDB) Close() error {
	m.db.Close()
	return nil
}

// -- Batch operations

func (m *MemDB) Batch() zerokv.Batch {
	return &MemBatch{batch: m.db.NewBatch()}
}

func (m *MemBatch) Put(key []byte, data []byte) error {
	return m.batch.Put(key, data)
}

// BatchDel adds a delete operation to the current batch.
func (m *MemBatch) Delete(key []byte) error {
	return m.batch.Delete(key)
}

// flushBatch flushes any pending batch operations.
func (m *MemBatch) Commit(ctx context.Context) error {
	return m.batch.Write()
}
func (m *MemDB) Scan(value []byte) zerokv.Iterator {
	// MemDB does not support range scanning/iteration
	// Returns nil to indicate this operation is not supported
	return nil
}
