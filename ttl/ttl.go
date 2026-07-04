package ttl

import (
	"context"
	"time"

	zerokv "github.com/rawbytedev/zerokv/core"
)

// TTLProvider is an optional interface for databases that support TTL
type TTLProvider interface {
	zerokv.Core
	GetTTL(ctx context.Context, key []byte) (time.Duration, error)
	PutTTL(ctx context.Context, key, data []byte, ttl time.Duration) error
	UpdateTTL(ctx context.Context, key []byte, ttl time.Duration) error
}

// BatchWithTTL is an optional extension for batch TTL support
type BatchWithTTL interface {
	zerokv.Batch
	PutTTL(key []byte, data []byte, ttl time.Duration) error
	UpdateTTL(key []byte, ttl time.Duration) error
}
