package fixture

import (
	"crypto/rand"
	"testing"

	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/implementations/badgerdb"
	"github.com/rawbytedev/zerokv/implementations/memdb"
	"github.com/rawbytedev/zerokv/implementations/pebbledb"
	encoders "github.com/rawbytedev/zerokv/serializers"
)

var AllDBs = []string{
	"pebbledb",
	"badgerdb",
	"memdb",
}

// setupBadgerDB creates a temporary BadgerDB instance for testing.
func SetupDB(t *testing.T, name string) zerokv.Core {
	tmp := t.TempDir()
	var db zerokv.Core
	var err error
	switch name {
	case "badgerdb":
		db, err = badgerdb.NewBadgerDB(badgerdb.Config{
			Dir: tmp,
		})
	case "pebbledb":
		db, err = pebbledb.NewPebbleDB(pebbledb.Config{
			Dir: tmp,
		})
	case "memdb":
		db, err = memdb.NewMemDB(memdb.Config{})
	}
	if err != nil || db == nil {
		t.Fatalf("Failed to create %s: %v", name, err)
	}
	return db
}

func SetupEncoders(t *testing.T, name string) encoders.Encoder {
	switch name {
	case "json":
		return encoders.NewJsonEncoder()
	case "rlp":
		return encoders.NewRLPEncoder()
	case "yaml":
		return encoders.NewYamlEncoder()
	}
	t.Fatalf("Failed to create %s encoder", name)
	return nil
}

// randomBytes generates a slice of random bytes of specified length.
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
