package pebbledb

import (
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
