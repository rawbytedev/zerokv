package pebbledb

import (
	"context"

	"github.com/cockroachdb/pebble"
)

// specific Pebbledb options
type Config struct {
	Dir           string
	PebbleConfigs *pebble.Options
}

func DefaultOptions(Dir string) *Config {
	return &Config{Dir, nil}
}

type writeOptsKey struct{}

// WithSync returns a context that forces the next write operation to use pebble.Sync.
func WithSync(ctx context.Context) context.Context {
	return context.WithValue(ctx, writeOptsKey{}, pebble.Sync)
}

// WithNoSync returns a context that forces pebble.NoSync.
func WithNoSync(ctx context.Context) context.Context {
	return context.WithValue(ctx, writeOptsKey{}, pebble.NoSync)
}

// getWriteOpts extracts the write options from context, defaulting to pebble.Sync.
func getWriteOpts(ctx context.Context) *pebble.WriteOptions {
	if val := ctx.Value(writeOptsKey{}); val != nil {
		if opts, ok := val.(*pebble.WriteOptions); ok {
			return opts
		}
	}
	return pebble.Sync // default for most users
}
