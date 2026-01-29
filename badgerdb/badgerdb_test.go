package badgerdb_test

import (
	"testing"

	"github.com/rawbytedev/zerokv/helpers"
	"github.com/stretchr/testify/require"
)

// TestBadgerBatchOperations tests batch Put and Get operations.
func TestBadgerBatchOperations(t *testing.T) {
	db := helpers.SetupDB(t, "badgerdb")
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
	// This should fail because the batch has already been committed
	err = batch.Put(keys[0], values[1])
	require.Error(t, err, "This transaction has been discarded. Create a new one")
	// This should also fail because the batch has already been committed
	err = batch.Commit(t.Context())
	require.Error(t, err, "Batch commit not permitted after finish")
	defer db.Close()
}
