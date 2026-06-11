package pebbledb_test

import (
	"bytes"
	"testing"

	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/implementations/pebbledb"
	"github.com/rawbytedev/zerokv/testing/fixture"
	"github.com/stretchr/testify/require"
)

// TestPebbleBatchOperations tests batch Put and Get operations.
func TestPebbleBatchOperations(t *testing.T) {
	db := fixture.SetupDB(t, "pebbledb")
	defer db.Close()
	batch := db.Batch()
	keys := make([][]byte, 5)
	values := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		keys[i] = fixture.RandomBytes(16)
		values[i] = fixture.RandomBytes(32)
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
	// Attempting to use the batch after commit should result in panic this behavior is specific to PebbleDB
	// This should fail because the batch has already been committed
	require.Panics(t, func() {
		batch.Put(keys[0], values[1])
	})
	// This should also fail because the batch has already been committed
	require.Panics(t, func() {
		batch.Commit(t.Context())
	})
}

// Helper to fill database with test values
func fillPebbleValues(t *testing.T, db zerokv.Core) ([][]byte, [][]byte) {
	keys := make([][]byte, 10)
	values := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = fixture.RandomBytes(16)
		values[i] = fixture.RandomBytes(32)
		pref_key := make([]byte, 0)
		pref_key = append(pref_key, []byte("pre_")...)
		pref_key = append(pref_key, keys[i]...)
		err := db.Put(t.Context(), pref_key, values[i])
		require.NoError(t, err)
	}
	return keys, values
}

// TestNil tests
func TestKeyNil(t *testing.T) {
	db := fixture.SetupDB(t, "pebbledb")
	defer db.Close()
	err := db.Put(t.Context(), nil, nil)
	require.NoError(t, err, "Pebble doesn't return error on key nil")
}

// TestPebbleReverseIterator tests full reverse iteration
func TestPebbleReverseIterator(t *testing.T) {
	db := fixture.SetupDB(t, "pebbledb")
	defer db.Close()

	pdb := db.(*pebbledb.PebbleDB)
	require.NotNil(t, pdb)

	_, values := fillPebbleValues(t, pdb)

	// Test full reverse iteration
	it := pebbledb.NewReverseIterator(pdb)
	defer it.Release()
	require.NotNil(t, it, "Reverse iterator should not be nil")

	count := 0
	for it.Next() {
		count++
		require.NotNil(t, it.Key(), "Key should not be nil")
		require.NotNil(t, it.Value(), "Value should not be nil")

		// Verify value is in our set
		found := false
		for i := 0; i < len(values); i++ {
			if bytes.Equal(values[i], it.Value()) {
				found = true
				break
			}
		}
		require.True(t, found, "Retrieved value should be from inserted values")
	}
	require.True(t, count == 10, "Should iterate 10 times")
	require.NoError(t, it.Error(), "Iterator should have no errors")
}

// TestPebbleReversePrefixIterator tests reverse iteration with prefix
func TestPebbleReversePrefixIterator(t *testing.T) {
	db := fixture.SetupDB(t, "pebbledb")
	defer db.Close()

	pdb := db.(*pebbledb.PebbleDB)
	require.NotNil(t, pdb)

	_, values := fillPebbleValues(t, pdb)

	// Test reverse prefix iteration
	it := pebbledb.NewReversePrefixIterator(pdb, []byte("pre_"))
	defer it.Release()
	require.NotNil(t, it, "Reverse prefix iterator should not be nil")

	count := 0
	for it.Next() {
		count++
		key := it.Key()
		require.NotNil(t, key, "Key should not be nil")
		require.True(t, bytes.HasPrefix(key, []byte("pre_")), "Key should have prefix")

		// Verify value is in our set
		found := false
		for i := 0; i < len(values); i++ {
			if bytes.Equal(values[i], it.Value()) {
				found = true
				break
			}
		}
		require.True(t, found, "Retrieved value should be from inserted values")
	}
	require.Equal(t, 10, count, "Should iterate exactly 10 times")
	require.NoError(t, it.Error(), "Iterator should have no errors")
}

// TestPebbleReverseIteratorOrder verifies reverse order
func TestPebbleReverseIteratorOrder(t *testing.T) {
	db := fixture.SetupDB(t, "pebbledb")
	defer db.Close()

	pdb := db.(*pebbledb.PebbleDB)
	require.NotNil(t, pdb)

	// Insert keys in predictable order
	keys := [][]byte{
		[]byte("key_01"),
		[]byte("key_02"),
		[]byte("key_03"),
		[]byte("key_04"),
		[]byte("key_05"),
	}
	for _, key := range keys {
		err := pdb.Put(t.Context(), key, []byte("value"))
		require.NoError(t, err)
	}

	// Get forward order
	forwardKeys := make([][]byte, 0)
	it := pdb.Scan([]byte("key_"))
	for it.Next() {
		forwardKeys = append(forwardKeys, it.Key())
	}
	it.Release()

	// Get reverse order
	reverseKeys := make([][]byte, 0)
	rit := pebbledb.NewReversePrefixIterator(pdb, []byte("key_"))
	defer rit.Release()
	for rit.Next() {
		reverseKeys = append(reverseKeys, rit.Key())
	}
	// Verify they're in opposite order
	require.Equal(t, len(forwardKeys), len(reverseKeys), "Should have same count")
	for i := 0; i < len(forwardKeys); i++ {
		require.Equal(t, forwardKeys[i], reverseKeys[len(reverseKeys)-1-i], "Keys should be in reverse order")
	}
}
