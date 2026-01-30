package pebbledb_test

import (
	"testing"

	"github.com/rawbytedev/zerokv/helpers"
	"github.com/stretchr/testify/require"
)

// TestPebbleBatchOperations tests batch Put and Get operations.
func TestPebbleBatchOperations(t *testing.T) {
	db := helpers.SetupDB(t, "pebbledb")
	batch := db.Batch()
	keys := make([][]byte, 5)
	values := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		keys[i] = helpers.RandomBytes(16)
		values[i] = helpers.RandomBytes(32)
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
	defer db.Close()
}
