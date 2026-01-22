package stores_test

import (
	"crypto/rand"
	"dbmodel"
	"dbmodel/configs"
	"dbmodel/stores"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupBadgerDB creates a temporary BadgerDB instance for testing.
func setupBadgerDB(t *testing.T) dbmodel.Store {
	tmp := t.TempDir()
	db, err := stores.NewBadgerDB(configs.StoreConfig{
		Default: &configs.DefaultOptions{
			Dir: tmp,
		},
	})
	if err != nil || db == nil {
		t.Fatalf("Failed to create BadgerDB: %v", err)
	}
	return db
}

// randomBytes generates a slice of random bytes of specified length.
func randomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// TestBadgerGetPutDelete tests basic Put, Get, and Delete operations.
func TestBadgerGetPutDelete(t *testing.T) {
	db := setupBadgerDB(t)
	keys := make([][]byte, 10)
	values := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = randomBytes(16)
		values[i] = randomBytes(32)
		err := db.Put(keys[i], values[i])
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}
	for i := 0; i < 10; i++ {
		value, err := db.Get(keys[i])
		require.NoError(t, err, "Error retrieving value for key")
		require.Equal(t, values[i], value, "Retrieved value does not match expected")
		err = db.Delete(keys[i])
		require.NoError(t, err, "Error deleting key")
		_, err = db.Get(keys[i])
		require.Error(t, err, "Expected error retrieving deleted key")
	}
	defer db.Close()
}

// TestBadgerGetNonExistentKey tests retrieval of a non-existent key.
func TestBadgerGetNonExistentKey(t *testing.T) {
	db := setupBadgerDB(t)
	nonExistentKey := randomBytes(16)
	_, err := db.Get(nonExistentKey)
	require.Error(t, err, "Expected error when getting non-existent key")
	defer db.Close()
}

// TestBadgerOverwriteKey tests overwriting an existing key.
func TestBadgerOverwriteKey(t *testing.T) {
	db := setupBadgerDB(t)
	key := randomBytes(16)
	value1 := randomBytes(32)
	value2 := randomBytes(32)
	err := db.Put(key, value1)
	require.NoError(t, err, "Error putting first value")
	retrievedValue, err := db.Get(key)
	require.NoError(t, err, "Error getting first value")
	require.Equal(t, value1, retrievedValue, "First retrieved value does not match")
	err = db.Put(key, value2)
	require.NoError(t, err, "Error putting second value")
	retrievedValue, err = db.Get(key)
	require.NoError(t, err, "Error getting second value")
	require.Equal(t, value2, retrievedValue, "Second retrieved value does not match")
	defer db.Close()
}

// TestBadgerClose tests closing the BadgerDB instance.
func TestBadgerClose(t *testing.T) {
	db := setupBadgerDB(t)
	err := db.Close()
	require.NoError(t, err, "Error closing BadgerDB")
}

// TestBadgerBatchOperations tests batch Put and Get operations.
func TestBadgerBatchOperations(t *testing.T) {
	db := setupBadgerDB(t)
	batch := db.Batch()
	keys := make([][]byte, 5)
	values := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		keys[i] = randomBytes(16)
		values[i] = randomBytes(32)
		err := batch.Put(keys[i], values[i])
		require.NoError(t, err, "Error adding Put operation to batch")
	}
	err := batch.Commits()
	require.NoError(t, err, "Error committing batch operations")
	for i := 0; i < 5; i++ {
		retrievedValue, err := db.Get(keys[i])
		require.NoError(t, err, "Error getting value after batch commit")
		require.Equal(t, values[i], retrievedValue, "Retrieved value does not match expected after batch commit")
	}
	// This should fail because the batch has already been committed
	err = batch.Put(keys[0], values[1])
	require.Error(t, err, "This transaction has been discarded. Create a new one")
	// This should also fail because the batch has already been committed
	err = batch.Commits()
	require.Error(t, err, "Batch commit not permitted after finish")
	defer db.Close()
}
