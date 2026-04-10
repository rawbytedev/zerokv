package badgerdb

import (
	"context"
	"errors"

	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/internal"

	"github.com/dgraph-io/badger/v4"
)

type BadgerDB struct {
	db *badger.DB
}
type badgerBatch struct {
	batch *badger.WriteBatch
}

type badgerIterator struct {
	Iterator *badger.Iterator
	started  bool
	valid    bool
	err      internal.IteratorErrors
}

// NewBadgerDB initializes and returns a zerokv.Core instance at the specified path(BadgerDB).
func NewBadgerDB(cfg Config) (zerokv.Core, error) {
	var opts badger.Options
	if cfg.BadgerConfigs != nil {
		opts = *cfg.BadgerConfigs
	} else {
		opts = badger.DefaultOptions(cfg.Dir)
	}
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerDB{db: db}, nil
}

// --- Basic CRUD operations ---

// Put inserts or updates a key-value pair in the database.
func (b *BadgerDB) Put(ctx context.Context, key, value []byte) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get retrieves the value for a given key. Returns an error if not found.
func (b *BadgerDB) Get(ctx context.Context, key []byte) ([]byte, error) {
	if err := internal.CheckContext(ctx); err != nil {
		return nil, err
	}
	var data []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = make([]byte, len(val))
			copy(data, val)
			return nil
		})
	})
	return data, err
}

// Delete removes a key-value pair from the database.
func (b *BadgerDB) Delete(ctx context.Context, key []byte) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close closes the BadgerDB instance and releases all resources.
func (b *BadgerDB) Close() error {
	var errs []error
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// -- Batch operations

// Batch creates a new batch operation for the BadgerDB instance.
func (b *BadgerDB) Batch() zerokv.Batch {
	return &badgerBatch{batch: b.db.NewWriteBatch()}
}

// Put inserts or updates a key-value pair in the batch.
func (b *badgerBatch) Put(key, value []byte) error {
	return b.batch.Set(key, value)
}

// Delete removes a key-value pair from the batch.
func (b *badgerBatch) Delete(key []byte) error {
	return b.batch.Delete(key)
}

// Commits commits the batch operations to the database.
func (b *badgerBatch) Commit(ctx context.Context) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return b.batch.Flush()
}

// -- Iterator operations

func (b *BadgerDB) Scan(prefix []byte) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: prefix, PrefetchValues: true})
	return &badgerIterator{Iterator: it}
}
func (it *badgerIterator) Next() bool {
	if !it.started {
		it.Iterator.Rewind()
		it.started = true
	} else {
		it.Iterator.Next()
	}
	it.valid = it.Iterator.Valid()
	return it.valid
}

func (it *badgerIterator) Key() []byte {
	if !it.valid {
		return nil
	}
	return it.Iterator.Item().KeyCopy(nil) // safer, doesn't make changes to key
}
func (it *badgerIterator) Value() []byte {
	if !it.valid {
		return nil
	}
	data, err := it.Iterator.Item().ValueCopy(nil)
	it.err.AddError(err)
	return data
}

// Release Must be called to avoid memory leaks
func (it *badgerIterator) Release() {
	it.Iterator.Close()
}

func (it *badgerIterator) Error() error {
	return it.err.Error()
}

//  --- specials methods to use with an instance of badgerdb for some other operations

func NewIterator(b *BadgerDB) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(badger.IteratorOptions{PrefetchValues: true})
	return &badgerIterator{Iterator: it}
}
func NewPrefixIterator(b *BadgerDB, prefix []byte) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: prefix, PrefetchValues: true})
	return &badgerIterator{Iterator: it}
}

type badgerReverseIterator struct {
	Iterator *badger.Iterator
	started  bool
	valid    bool
	err      internal.IteratorErrors
}

func (it *badgerReverseIterator) Next() bool {
	if !it.started {
		it.Iterator.Seek([]byte{0xFF}) // Start from the end of the keyspace
		it.started = true
	} else {
		it.Iterator.Next()
	}
	it.valid = it.Iterator.Valid()
	return it.valid
}

func (it *badgerReverseIterator) Key() []byte {
	if !it.valid {
		return nil
	}
	return it.Iterator.Item().KeyCopy(nil) // safer, doesn't make changes to key
}
func (it *badgerReverseIterator) Value() []byte {
	if !it.valid {
		return nil
	}
	data, err := it.Iterator.Item().ValueCopy(nil)
	it.err.AddError(err)
	return data
}

// Release Must be called to avoid memory leaks
func (it *badgerReverseIterator) Release() {
	it.Iterator.Close()
}

func (it *badgerReverseIterator) Error() error {
	return it.err.Error()
}

func NewReverseIterator(b *BadgerDB) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(badger.IteratorOptions{Reverse: true, PrefetchValues: true,
		PrefetchSize: 100})
	return &badgerReverseIterator{Iterator: it}
}

func NewReversePrefixIterator(b *BadgerDB, prefix []byte) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: []byte(prefix), PrefetchValues: true, PrefetchSize: 100, Reverse: true})
	return &badgerReverseIterator{Iterator: it}
}
