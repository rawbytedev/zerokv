package badgerdb

import (
	"context"
	"errors"

	"github.com/rawbytedev/zerokv"

	"github.com/dgraph-io/badger/v4"
)

type badgerDB struct {
	db *badger.DB
}
type badgerBatch struct {
	batch *badger.WriteBatch
}

type badgerIteractor struct {
	iteractor *badger.Iterator
	started   bool
	valid     bool
	err       []error
}

// NewBadgerDB initializes and returns a zerokv.Core instance at the specified path(badgerDB).
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
	return &badgerDB{db: db}, nil
}

// --- Basic CRUD operations ---

// Put inserts or updates a key-value pair in the database.
func (b *badgerDB) Put(ctx context.Context, key, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get retrieves the value for a given key. Returns an error if not found.
func (b *badgerDB) Get(ctx context.Context, key []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var data []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
	})
	return data, err
}

// Delete removes a key-value pair from the database.
func (b *badgerDB) Delete(ctx context.Context, key []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close closes the BadgerDB instance and releases all resources.
func (b *badgerDB) Close() error {
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
func (b *badgerDB) Batch() zerokv.Batch {
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
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.batch.Flush()
}

// -- Iterator operations

func (b *badgerDB) Scan(prefix []byte) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: prefix, PrefetchValues: true})
	return &badgerIteractor{iteractor: it}
}
func (it *badgerIteractor) Next() bool {
	if !it.started {
		it.iteractor.Rewind()
		it.started = true
	} else {
		it.iteractor.Next()
	}
	it.valid = it.iteractor.Valid()
	return it.valid
}

func (it *badgerIteractor) Key() []byte {
	if !it.valid {
		return nil
	}
	return it.iteractor.Item().KeyCopy(nil) // safer, doesn't make changes to key
}
func (it *badgerIteractor) Value() []byte {
	if !it.valid {
		return nil
	}
	data, err := it.iteractor.Item().ValueCopy(nil)
	if err != nil {
		it.err = append(it.err, err)
		return []byte{}
	}
	return data
}
func (it *badgerIteractor) Release() {
	it.iteractor.Close()
}
func (it *badgerIteractor) Error() error {
	return it.err[len(it.err)-1] // returns the most recent error
}

//  --- specials methods to use with an instance of badgerdb for some other operations

func NewIterator(b *badgerDB) zerokv.Iterator {
	it := &badger.Iterator{}
	b.db.View(func(txn *badger.Txn) error {
		it = txn.NewIterator(badger.IteratorOptions{})
		return nil
	})
	return &badgerIteractor{iteractor: it}
}
func NewReverseIterator(b *badgerDB) zerokv.Iterator {
	it := &badger.Iterator{}
	b.db.View(func(txn *badger.Txn) error {
		it = txn.NewIterator(badger.IteratorOptions{Reverse: true})
		return nil
	})
	return &badgerIteractor{iteractor: it}
}
func NewPrefixIterator(b *badgerDB, prefix []byte) zerokv.Iterator {
	it := &badger.Iterator{}
	b.db.View(func(txn *badger.Txn) error {
		it = txn.NewIterator(badger.IteratorOptions{Prefix: prefix})
		return nil
	})
	return &badgerIteractor{iteractor: it}
}
func NewReversePrefixIterator(b *badgerDB, prefix []byte) zerokv.Iterator {
	it := &badger.Iterator{}
	b.db.View(func(txn *badger.Txn) error {
		it = txn.NewIterator(badger.IteratorOptions{Prefix: prefix, Reverse: true})
		return nil
	})
	return &badgerIteractor{iteractor: it}
}
