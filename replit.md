# ZeroKV

A minimal, zero-overhead key-value store abstraction layer for Go. Provides a common interface for different key-value database backends (BadgerDB, PebbleDB, MemDB).

## Project Structure

- `interface.go` - Core interfaces: `Core`, `Batch`, `Iterator`
- `badgerdb/` - BadgerDB backend implementation
- `pebbledb/` - PebbleDB backend implementation  
- `memdb/` - In-memory backend implementation
- `encoders/` - JSON, RLP, and YAML serialization
- `tests/` - Shared integration tests across all backends
- `helpers/` - Test utilities and context helpers
- `examples/` - Usage examples

## Tech Stack

- **Language:** Go 1.25.2
- **Package Manager:** Go Modules (`go.mod` / `go.sum`)
- **Database Backends:** BadgerDB v4, PebbleDB, MemDB
- **Testing:** `github.com/stretchr/testify`

## Running Tests

```bash
go test ./... -v
```

## Known Issues

- `TestPebbleReverseIteratorOrder` in `pebbledb/` fails — this is a pre-existing bug in the PebbleDB reverse iterator implementation present in the imported repository. The reverse iterator does not correctly return keys in descending order.

## Workflow

The "Start application" workflow runs `go test ./... -v -count=1` to verify the library builds and all tests run.
