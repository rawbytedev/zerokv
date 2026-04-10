package memdb

// Config holds the configuration options for initializing the MemDB instance.
type MemConfig struct {
	size uint64
}

// specific Pebbledb options
type Config struct {
	Dir        string
	MemConfigs *MemConfig // this doesn't support options for now, but we can add it in the future if needed
}

func DefaultOptions(Dir string) *Config {
	return &Config{Dir, nil}
}
