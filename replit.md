# ZeroKV

A minimal, zero-overhead key-value store abstraction layer for Go. Provides a common interface for different key-value database backends (BadgerDB, PebbleDB, MemDB).

## Project Structure

- `interface.go` — Core interfaces: `Core`, `Batch`, `Iterator`
- `badgerdb/` — BadgerDB backend implementation
- `pebbledb/` — PebbleDB backend implementation
- `memdb/` — In-memory backend implementation
- `encoders/` — JSON, RLP, and YAML serialization
- `tests/` — Shared integration tests across all backends
- `helpers/` — Test utilities and context helpers
- `examples/` — Usage examples
- `cmd/demo/` — Interactive web demo server (Go HTTP + static HTML)
  - `main.go` — HTTP server exposing REST API over ZeroKV, seeds all 3 backends
  - `static/index.html` — Rich single-page UI for live interaction

## Tech Stack

- **Language:** Go 1.25.2
- **Package Manager:** Go Modules (`go.mod` / `go.sum`)
- **Database Backends:** BadgerDB v4, PebbleDB, MemDB
- **Testing:** `github.com/stretchr/testify`
- **Demo UI:** Vanilla HTML/CSS/JS (no framework), served by Go's net/http

## Running the Demo

The workflow runs `go run ./cmd/demo` which starts an HTTP server on port 5000.

The demo allows users to:
- Switch between all 3 backends (MemDB, BadgerDB, PebbleDB) live
- Perform Put, Get, Delete operations
- Prefix-scan / iterate over keys
- Commit atomic batch writes
- See a live database view and activity log

## Running Tests

```bash
go test ./... -v
```

## Known Issues

- `TestPebbleReverseIteratorOrder` in `pebbledb/` fails — pre-existing bug in the imported repo's PebbleDB reverse iterator implementation.
