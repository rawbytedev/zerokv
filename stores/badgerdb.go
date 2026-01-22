package stores

import (
	"dbmodel"
	dbconfig "dbmodel/configs"
	"errors"

	"github.com/dgraph-io/badger/v4"
)

type badgerdb struct {
	db *badger.DB
}
type badgerBatch struct {
	batch *badger.WriteBatch
}

// NewBadgerDB initializes and returns a BadgerDB instance at the specified path.
func NewBadgerDB(cfg dbconfig.StoreConfig) (dbmodel.Store, error) {
	var opts badger.Options
	if cfg.BadgerConfigs != nil {
		opts = *cfg.BadgerConfigs
	} else {
		opts = badger.DefaultOptions(cfg.Default.Dir)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &badgerdb{db: db}, nil
}

// Put inserts or updates a key-value pair in the database.
func (b *badgerdb) Put(key, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get retrieves the value for a given key. Returns an error if not found.
func (b *badgerdb) Get(key []byte) ([]byte, error) {
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
func (b *badgerdb) Delete(key []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close closes the BadgerDB instance and releases all resources.
func (b *badgerdb) Close() error {
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

// Batch creates a new batch operation for the BadgerDB instance.
/*
Must be used carefully calling Batch creates a new write batch that needs to be committed separately.
or else it may lead to uncommitted data. and data loss.
*/
func (b *badgerdb) Batch() dbmodel.Batch {
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
/*
Note: It is crucial to call Commits to ensure that all batched operations are saved to the database.
and no new insertions/updates/deletions will be saved until Commits is called.
Inserting to already committed batch is forbidden and will lead to errors.
*/
func (b *badgerBatch) Commits() error {
	return b.batch.Flush()
}
func (b *badgerdb) Scan(prefix []byte) dbmodel.Iterator {
	// Placeholder for Scan operation implementation
	return nil
}
func (b *badgerdb) NewIterator() dbmodel.Iterator {
	// Placeholder for Iterator implementation
	return nil
}
func NewReverseIterator() dbmodel.Iterator {
	// Placeholder for Reverse Iterator implementation
	return nil
}
func NewPrefixIterator(prefix []byte) dbmodel.Iterator {
	// Placeholder for Prefix Iterator implementation
	return nil
}
func NewReversePrefixIterator(prefix []byte) dbmodel.Iterator {
	// Placeholder for Reverse Prefix Iterator implementation
	return nil
}
