package ttl

import (
	"context"
	"time"

	zerokv "github.com/rawbytedev/zerokv/core"
)

// TTLProvider is an optional interface for databases that support TTL
type TTLProvider interface {
	TTLPut(ctx context.Context, key, data []byte, ttl time.Duration) error
}

// BatchWithTTL is an optional extension for batch TTL support
type BatchWithTTL interface {
	zerokv.Batch
	PutWithTTL(key []byte, data []byte, ttl time.Duration) error
}
