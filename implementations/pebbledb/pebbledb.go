package pebbledb

import (
	"context"
	"errors"

	"github.com/cockroachdb/pebble"
	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/internal"
)

type PebbleDB struct {
	db *pebble.DB
}
type pebbleBatch struct {
	batch *pebble.Batch
}
type pebbleIterator struct {
	Iterator *pebble.Iterator
	started  bool
	valid    bool
	err      internal.IteratorErrors
}

type pebbleReverseIterator struct {
	Iterator *pebble.Iterator
	started  bool
	valid    bool
	err      internal.IteratorErrors
}

// NewPebbleDB initializes and returns a zerokv.Core instance at the specified path(PebbleDB).
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
	return &PebbleDB{db: db}, nil
}

// --- Basic CRUD operations ---

// Put inserts or updates a key-value pair in the database.
func (p *PebbleDB) Put(ctx context.Context, key []byte, data []byte) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return p.db.Set(key, data, pebble.Sync)
}

// Get retrieves the value for a given key. Returns an error if not found.
func (p *PebbleDB) Get(ctx context.Context, key []byte) ([]byte, error) {
	if err := internal.CheckContext(ctx); err != nil {
		return nil, err
	}
	val, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return val, nil
}

// Del deletes a key-value pair from the database.
func (p *PebbleDB) Delete(ctx context.Context, key []byte) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return p.db.Delete(key, pebble.Sync)
}

// Close closes the database and releases all resources.
func (p *PebbleDB) Close() error {
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

func (p *PebbleDB) Batch() zerokv.Batch {
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
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	return p.batch.Commit(pebble.Sync)
}

// -- Iterator operations

func (p *PebbleDB) Scan(prefix []byte) zerokv.Iterator {
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
	return &pebbleIterator{Iterator: it, valid: false, started: false}
}

func (it *pebbleIterator) Next() bool {
	// this comes from how iterators works in pebble
	if !it.started {
		it.valid = it.Iterator.First()
		it.started = true
	} else {
		it.valid = it.Iterator.Next()
	}
	return it.valid
}

func (it *pebbleIterator) Key() []byte {
	if !it.valid {
		return nil
	}
	return it.Iterator.Key() // safer, doesn't make changes to key
}
func (it *pebbleIterator) Value() []byte {
	if !it.valid {
		return nil
	}
	data, err := it.Iterator.ValueAndErr()
	it.err.AddError(err)
	return data
}
func (it *pebbleIterator) Release() {
	it.valid = false
	it.Iterator.Close()
}
func (it *pebbleIterator) Error() error {
	return it.err.Error()
}

// --- specials methods to use with an instance of badgerdb for some other operations
func NewIterator(p *PebbleDB) zerokv.Iterator {
	it, err := p.db.NewIter(&pebble.IterOptions{})

	if err != nil {
		return nil
	}
	return &pebbleIterator{Iterator: it, valid: false, started: false}
}

func NewPrefixIterator(p *PebbleDB, prefix []byte) zerokv.Iterator {
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
	return &pebbleIterator{Iterator: it, valid: false, started: false}
}

// --- Reverse Iterators ---

func NewReverseIterator(p *PebbleDB) zerokv.Iterator {
	it, err := p.db.NewIter(&pebble.IterOptions{})
	if err != nil {
		return nil
	}
	return &pebbleReverseIterator{Iterator: it, valid: false, started: false}
}

func NewReversePrefixIterator(p *PebbleDB, prefix []byte) zerokv.Iterator {
	upbound := make([]byte, len(prefix))
	copy(upbound, prefix)
	if len(upbound) > 0 {
		upbound[len(upbound)-1]++
	}
	it, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: upbound,
	})
	if err != nil {
		return nil
	}
	return &pebbleReverseIterator{Iterator: it, valid: false, started: false}
}

func (it *pebbleReverseIterator) Next() bool {
	if !it.started {
		it.valid = it.Iterator.Last()
		it.started = true
	} else {
		it.valid = it.Iterator.Prev()
	}
	return it.valid
}

func (it *pebbleReverseIterator) Key() []byte {
	if !it.valid {
		return nil
	}
	return it.Iterator.Key()
}

func (it *pebbleReverseIterator) Value() []byte {
	if !it.valid {
		return nil
	}
	data, err := it.Iterator.ValueAndErr()
	it.err.AddError(err)
	return data
}

func (it *pebbleReverseIterator) Release() {
	it.valid = false
	it.Iterator.Close()
}

func (it *pebbleReverseIterator) Error() error {
	return it.err.Error()
}
