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
	reverse  bool
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
	return p.db.Set(key, data, getWriteOpts(ctx))
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
	return p.db.Delete(key, getWriteOpts(ctx))
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
	return p.batch.Set(key, data, nil)
}

// BatchDel adds a delete operation to the current batch.
func (p *pebbleBatch) Delete(key []byte) error {
	return p.batch.Delete(key, nil)
}

// flushBatch flushes any pending batch operations.
func (p *pebbleBatch) Commit(ctx context.Context) error {
	if err := internal.CheckContext(ctx); err != nil {
		return err
	}
	opts := getWriteOpts(ctx)
	return p.batch.Commit(opts)
}

// -- Iterator operations

func (p *PebbleDB) Scan(prefix []byte, opts ...zerokv.ScanOption) zerokv.Iterator {
	scanCfg := zerokv.NewScanConfig()
	for _, opt := range opts {
		opt(scanCfg)
	}
	var upbound []byte
	if prefix != nil {
		upbound = make([]byte, len(prefix))
		copy(upbound, prefix)
		upbound[len(upbound)-1]++
	}
	it, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: upbound,
	})
	if err != nil {
		return nil

	}

	return &pebbleIterator{Iterator: it, valid: false, started: false, reverse: scanCfg.Reverse}
}

func (it *pebbleIterator) Next() bool {
	switch it.started {
	case true:
		if it.reverse {
			it.valid = it.Iterator.Prev()
		} else {
			it.valid = it.Iterator.Next()
		}
	case false:
		if it.reverse {
			it.valid = it.Iterator.Last()
		} else {
			it.valid = it.Iterator.First()
		}
		it.started = true
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

func (it *pebbleIterator) Reset() {
	it.Iterator.First()
}
func (it *pebbleIterator) Error() error {
	return it.err.Error()
}
