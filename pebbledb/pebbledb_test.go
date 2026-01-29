package pebbledb_test

import (
	"crypto/rand"
	"testing"

	"github.com/rawbytedev/zerokv"

	"github.com/rawbytedev/zerokv/pebbledb"
	"github.com/stretchr/testify/require"
)

// setupPebbleDB creates a temporary PebbleDB instance for testing.
func setupPebbleDB(t *testing.T) zerokv.Core {
	tmp := t.TempDir()
	db, err := pebbledb.NewPebbleDB(pebbledb.Config{
		Dir: tmp,
	})
	if err != nil || db == nil {
		t.Fatalf("Failed to create PebbleDB: %v", err)
	}
	return db
}

// randomBytes generates a slice of random bytes of specified length.
func randomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// TestPebbleGetPutDelete tests basic Put, Get, and Delete operations.
func TestGetPutDelete(t *testing.T) {
	db := setupPebbleDB(t)
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

// TestPebbleGetNonExistentKey tests retrieval of a non-existent key.
func TestPebbleGetNonExistentKey(t *testing.T) {
	db := setupPebbleDB(t)
	nonExistentKey := randomBytes(16)
	_, err := db.Get(t.Context(), nonExistentKey)
	require.Error(t, err, "Expected error when getting non-existent key")
	defer db.Close()
}

// TestPebbleOverwriteKey tests overwriting an existing key.
func TestPebbleOverwriteKey(t *testing.T) {
	db := setupPebbleDB(t)
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

// TestPebbleClose tests closing the PebbleDB instance.
func TestPebbleClose(t *testing.T) {
	db := setupPebbleDB(t)
	err := db.Close()
	require.NoError(t, err, "Error closing PebbleDB")
}

// TestPebbleBatchOperations tests batch Put and Get operations.
func TestPebbleBatchOperations(t *testing.T) {
	db := setupPebbleDB(t)
	batch := db.Batch()
	keys := make([][]byte, 5)
	values := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		keys[i] = randomBytes(16)
		values[i] = randomBytes(32)
		err := batch.Put(keys[i], values[i])
		require.NoError(t, err, "Error adding Put operation to batch")
	}
	err := batch.Commit(t.Context())
	require.NoError(t, err, "Error committing batch operations")
	for i := 0; i < 5; i++ {
		retrievedValue, err := db.Get(t.Context(), keys[i])
		require.NoError(t, err, "Error getting value after batch commit")
		require.Equal(t, values[i], retrievedValue, "Retrieved value does not match expected after batch commit")
	}
	// This should fail because the batch has already been committed
	require.Panics(t, func() {
		batch.Put(keys[0], values[1])
	})
	// This should also fail because the batch has already been committed
	require.Panics(t, func() {
		batch.Commit(t.Context())
	})
	defer db.Close()
}
