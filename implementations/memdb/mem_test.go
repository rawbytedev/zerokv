package memdb_test

import (
	"testing"

	"github.com/rawbytedev/zerokv/testing/fixture"
	"github.com/stretchr/testify/require"
)

// TestMemdbBatchOperations tests batch Put and Get operations.
func TestMemdbBatchOperations(t *testing.T) {
	db := fixture.SetupDB(t, "memdb")
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
	defer db.Close()
}
