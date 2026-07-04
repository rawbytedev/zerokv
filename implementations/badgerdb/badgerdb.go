package badgerdb

import (
	"context"
	"errors"
	"time"

	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/internal"

	"github.com/dgraph-io/badger/v4"
)

type BadgerDB struct {
	db *badger.DB
}
type badgerBatch struct {
	batch *badger.WriteBatch
	db    *BadgerDB
}

type badgerIterator struct {
	Iterator *badger.Iterator
	started  bool
	valid    bool
	err      internal.IteratorErrors
	reverse  bool
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

// PutTTL inserts or updates a key-value pair in the database with a ttl attached to it
func (b *BadgerDB) PutTTL(ctx context.Context, key []byte, value []byte, duration time.Duration) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return b.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value).WithTTL(duration)
		return txn.SetEntry(e)
	})
}

func (b *BadgerDB) GetTTL(ctx context.Context, key []byte) (time.Duration, error) {
	var expires time.Duration
	if err := internal.CheckContext(ctx); err != nil {
		return 0, err
	}
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		expires = time.Until(time.Unix(int64(item.ExpiresAt()), 0))
		return nil
	})
	return expires, err
}

func (b *BadgerDB) UpdateTTL(ctx context.Context, key []byte, duration time.Duration) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return b.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			e := badger.NewEntry(key, val).WithTTL(duration)
			return txn.SetEntry(e)
		})
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
	return &badgerBatch{batch: b.db.NewWriteBatch(), db: b}
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

func (b *badgerBatch) PutTTL(key []byte, value []byte, ttl time.Duration) error {
	e := badger.NewEntry(key, value).WithTTL(ttl)
	return b.batch.SetEntry(e)
}

func (b *badgerBatch) UpdateTTL(key []byte, ttl time.Duration) error {
	value, err := b.db.Get(context.Background(), key)
	if err != nil {
		return err
	}
	e := badger.NewEntry(key, value).WithTTL(ttl)
	return b.batch.SetEntry(e)
}

// -- Iterator operations

func (b *BadgerDB) Scan(prefix []byte, opts ...zerokv.ScanOption) zerokv.Iterator {
	txn := b.db.NewTransaction(false)
	scanCfg := zerokv.NewScanConfig()
	for _, opt := range opts {
		opt(scanCfg)
	}
	var it *badger.Iterator
	if prefix != nil {
		it = txn.NewIterator(badger.IteratorOptions{Prefix: prefix, PrefetchValues: scanCfg.Prefetch, Reverse: scanCfg.Reverse})
	} else {
		it = txn.NewIterator(badger.IteratorOptions{PrefetchValues: scanCfg.Prefetch, Reverse: scanCfg.Reverse})
	}
	return &badgerIterator{Iterator: it, reverse: scanCfg.Reverse}
}

func (it *badgerIterator) Next() bool {
	if !it.started {
		if it.reverse {
			it.Iterator.Seek([]byte{0xFF})
		} else {
			it.Iterator.Rewind()
		}
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

func (it *badgerIterator) Reset() {
	it.Iterator.Rewind()
}

// Release Must be called to avoid memory leaks
func (it *badgerIterator) Release() {
	it.Iterator.Close()
}

func (it *badgerIterator) Error() error {
	return it.err.Error()
}
