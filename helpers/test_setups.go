package helpers

import (
	"crypto/rand"
	"testing"

	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/pebbledb"
)

// setupBadgerDB creates a temporary BadgerDB instance for testing.
func SetupDB(t *testing.T, name string) zerokv.Core {
	tmp := t.TempDir()
	var db zerokv.Core
	var err error
	if name == "badgerdb" {
		db, err = badgerdb.NewBadgerDB(badgerdb.Config{
			Dir: tmp,
		})
	} else {
		db, err = pebbledb.NewPebbleDB(pebbledb.Config{
			Dir: tmp,
		})
	}
	if err != nil || db == nil {
		t.Fatalf("Failed to create %s: %v", name, err)
	}
	return db
}

// randomBytes generates a slice of random bytes of specified length.
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
