package badgerdb

import "github.com/dgraph-io/badger/v4"

// specific badgerdb options
type Config struct {
	Dir           string
	BadgerConfigs *badger.Options
}

func DefaultOptions(Dir string) *Config {
	return &Config{Dir, nil}
}
