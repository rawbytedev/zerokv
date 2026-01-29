package pebbledb

import (
	"context"
	"errors"

	"github.com/cockroachdb/pebble"
	"github.com/rawbytedev/zerokv"
)

type pebbleDB struct {
	db *pebble.DB
}
type pebbleBatch struct {
	batch *pebble.Batch
}
type pebbleIteractor struct {
	iteractor *pebble.Iterator
	started   bool
	valid     bool
	err       []error
}

// NewPebbleDB initializes and returns a zerokv.Core instance at the specified path(pebbleDB).
func NewPebbleDB(cfg Config) (zerokv.Core, error) {
	opts := &pebble.Options{}
	if cfg.PebbleConfigs != nil {
		opts = cfg.PebbleConfigs
	} else {
		opts = &pebble.Options{}
	}
	db, err := pebble.Open(cfg.Dir, opts)
	if err != nil {
		return nil, err
	}
	return &pebbleDB{db: db}, nil
}

// --- Basic CRUD operations ---

// Put inserts or updates a key-value pair in the database.
func (p *pebbleDB) Put(ctx context.Context, key []byte, data []byte) error {
	return p.db.Set(key, data, pebble.Sync)
}

// Get retrieves the value for a given key. Returns an error if not found.
func (p *pebbleDB) Get(ctx context.Context, key []byte) ([]byte, error) {
	val, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return val, nil
}

// Del deletes a key-value pair from the database.
func (p *pebbleDB) Delete(ctx context.Context, key []byte) error {
	return p.db.Delete(key, pebble.Sync)
}

// Close closes the database and releases all resources.
func (p *pebbleDB) Close() error {
	var errs []error
	if err := p.db.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// -- Batch operations

func (p *pebbleDB) Batch() zerokv.Batch {
	return &pebbleBatch{batch: p.db.NewBatch()}
}

func (p *pebbleBatch) Put(key []byte, data []byte) error {
	return p.batch.Set(key, data, pebble.NoSync)
}

// BatchDel adds a delete operation to the current batch.
func (p *pebbleBatch) Delete(key []byte) error {
	return p.batch.Delete(key, pebble.NoSync)
}

// flushBatch flushes any pending batch operations.
func (p *pebbleBatch) Commit(ctx context.Context) error {
	return p.batch.Commit(pebble.Sync)
}

// -- Iterator operations

func (p *pebbleDB) Scan(prefix []byte) zerokv.Iterator {
	upbound := make([]byte, len(prefix))
	copy(upbound, prefix)
	upbound[len(upbound)-1]++
	it, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: upbound,
	})
	if err != nil {
		return nil
	}
	return &pebbleIteractor{iteractor: it, valid: false, started: false}
}

func (it *pebbleIteractor) Next() bool {
	// this comes from how iterators works in pebble
	if !it.started {
		it.valid = it.iteractor.First()
		it.started = true
	} else {
		it.valid = it.iteractor.Next()
	}
	return it.valid
}

func (it *pebbleIteractor) Key() []byte {
	if !it.valid {
		return nil
	}
	return it.iteractor.Key() // safer, doesn't make changes to key
}
func (it *pebbleIteractor) Value() []byte {
	if !it.valid {
		return nil
	}
	data, err := it.iteractor.ValueAndErr()
	if err != nil {
		it.err = append(it.err, err)
		return nil
	}
	return data
}
func (it *pebbleIteractor) Release() {
	it.valid = false
	it.iteractor.Close()
}
func (it *pebbleIteractor) Error() error {
	return it.err[len(it.err)-1] // returns the most recent error
}

//  --- specials methods to use with an instance of badgerdb for some other operations
func NewIterator(p *pebbleDB) zerokv.Iterator {
	it, err := p.db.NewIter(&pebble.IterOptions{})

	if err != nil {
		return nil
	}
	return &pebbleIteractor{iteractor: it, valid: false, started: false}
}

func NewPrefixIterator(p *pebbleDB, prefix []byte) zerokv.Iterator {
	upbound := make([]byte, len(prefix))
	copy(upbound, prefix)
	upbound[len(upbound)-1]++
	it, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: upbound,
	})
	if err != nil {
		return nil
	}
	return &pebbleIteractor{iteractor: it, valid: false, started: false}
}

/*
Due to how pebble works reverse Iterators have to be built using a different struct mainly because
we need to make use
it.Prev()
it.Last()
in Next()
*/

func NewReverseIterator(p *pebbleDB, prefix []byte) zerokv.Iterator {
	return nil
}
func NewReversePrefixIterator(prefix []byte) zerokv.Iterator {
	// Placeholder for Reverse Prefix Iterator implementation
	return nil
}
