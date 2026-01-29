package tests

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/pebbledb"
	"github.com/stretchr/testify/require"
)

type test struct {
	name string
	fn   func(t *testing.T, name string)
}

// setupBadgerDB creates a temporary BadgerDB instance for testing.
func setupDB(t *testing.T, name string) zerokv.Core {
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
func randomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
func TestZeroKvImplementation(t *testing.T) {
	dbs := []string{"badgerdb", "pebbledb"}
	list_test := []test{
		{name: "TestGetPutDelete",
			fn: func(t *testing.T, name string) {
				testGetPutDelete(t, name)
			}}, {
			name: "testGetNonExistentKey",
			fn: func(t *testing.T, name string) {
				testGetNonExistentKey(t, name)
			}}, {
			name: "TestOverwriteKey",
			fn: func(t *testing.T, name string) {
				testOverwriteKey(t, name)
			}},
		{
			name: "TestClose",
			fn: func(t *testing.T, name string) {
				testClose(t, name)
			}},
	}

	for i := range dbs {
		for tt := range list_test {
			testname := fmt.Sprintf("%s%s", list_test[tt].name, dbs[i])
			t.Run(testname, func(t *testing.T) {
				list_test[tt].fn(t, dbs[i])
			})
		}
	}

}

// TestGetPutDelete tests basic Put, Get, and Delete operations.
func testGetPutDelete(t *testing.T, name string) {
	db := setupDB(t, name)
	keys := make([][]byte, 10)
	values := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = randomBytes(16)
		values[i] = randomBytes(32)
		err := db.Put(t.Context(), keys[i], values[i])
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}
	for i := 0; i < 10; i++ {
		value, err := db.Get(t.Context(), keys[i])
		require.NoError(t, err, "Error retrieving value for key")
		require.Equal(t, values[i], value, "Retrieved value does not match expected")
		err = db.Delete(t.Context(), keys[i])
		require.NoError(t, err, "Error deleting key")
		_, err = db.Get(t.Context(), keys[i])
		require.Error(t, err, "Expected error retrieving deleted key")
	}
	defer db.Close()
}

// TestGetNonExistentKey tests retrieval of a non-existent key.
func testGetNonExistentKey(t *testing.T, name string) {
	db := setupDB(t, name)
	nonExistentKey := randomBytes(16)
	_, err := db.Get(t.Context(), nonExistentKey)
	require.Error(t, err, "Expected error when getting non-existent key")
	defer db.Close()
}

// TestOverwriteKey tests overwriting an existing key.
func testOverwriteKey(t *testing.T, name string) {
	db := setupDB(t, name)
	key := randomBytes(16)
	value1 := randomBytes(32)
	value2 := randomBytes(32)
	err := db.Put(t.Context(), key, value1)
	require.NoError(t, err, "Error putting first value")
	retrievedValue, err := db.Get(t.Context(), key)
	require.NoError(t, err, "Error getting first value")
	require.Equal(t, value1, retrievedValue, "First retrieved value does not match")
	err = db.Put(t.Context(), key, value2)
	require.NoError(t, err, "Error putting second value")
	retrievedValue, err = db.Get(t.Context(), key)
	require.NoError(t, err, "Error getting second value")
	require.Equal(t, value2, retrievedValue, "Second retrieved value does not match")
	defer db.Close()
}

// TestClose tests closing the PebbleDB instance.
func testClose(t *testing.T, name string) {
	db := setupDB(t, name)
	err := db.Close()
	require.NoError(t, err, "Error closing PebbleDB")
}
