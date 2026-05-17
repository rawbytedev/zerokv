package badgerdb_test

import (
	"bytes"
	"testing"
	"time"

	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/implementations/badgerdb"
	"github.com/rawbytedev/zerokv/testing/fixture"
	ttl "github.com/rawbytedev/zerokv/ttl"
	"github.com/stretchr/testify/require"
)

func TestTTTL(t *testing.T) {
	db := fixture.SetupDB(t, "badgerdb")
	defer db.Close()
	key := []byte{0x1, 0x12, 0x23}
	value := []byte{0x1, 0x12, 0x23}
	if ttlDB, ok := db.(ttl.TTLProvider); ok {
		ttlDB.TTLPut(t.Context(), key, value, 5*time.Second)
	} else {
		t.Fatal("this database does not support TTL")
	}
	time.Sleep(6 * time.Second)
	data, err := db.Get(t.Context(), key)
	require.Nil(t, data, "Value not nil after expiry")
	require.NotNil(t, err, "No error returned for key not found")
}

func TestBatchTTL(t *testing.T) {
	db := fixture.SetupDB(t, "badgerdb")
	defer db.Close()
	key := []byte{0x1, 0x12, 0x23}
	value := []byte{0x1, 0x12, 0x23}
	if ttlDB, ok := db.(ttl.TTLProvider); ok {
		ttlDB.TTLPut(t.Context(), key, value, 5*time.Second)
	} else {
		t.Fatal("this database does not support TTL")
	}
	time.Sleep(6 * time.Second)
	data, err := db.Get(t.Context(), key)
	require.Nil(t, data, "Value not nil after expiry")
	require.NotNil(t, err, "No error returned for key not found")
}

// TestBadgerBatchOperations tests batch Put and Get operations.
func TestBadgerBatchOperations(t *testing.T) {
	db := fixture.SetupDB(t, "badgerdb")
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
	// This should fail because the batch has already been committed
	err = batch.Put(keys[0], values[1])
	require.Error(t, err, "This transaction has been discarded. Create a new one")
	// This should also fail because the batch has already been committed
	err = batch.Commit(t.Context())
	require.Error(t, err, "Batch commit not permitted after finish")

}

// Helper to fill database with test values
func fillBadgerValues(t *testing.T, db zerokv.Core) ([][]byte, [][]byte) {
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

// TestBadgerReverseIterator tests full reverse iteration
func TestBadgerReverseIterator(t *testing.T) {
	// Create a temporary badgerdb instance directly
	tmp := t.TempDir()
	dbInterface, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: tmp})
	require.NoError(t, err)

	// Type assertion to get concrete type - this works within the package
	bdb := dbInterface.(*badgerdb.BadgerDB)
	require.NotNil(t, bdb)

	defer bdb.Close()
	_, values := fillBadgerValues(t, bdb)

	// Test full reverse iteration
	it := badgerdb.NewReverseIterator(bdb)
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

// TestBadgerReversePrefixIterator tests reverse iteration with prefix
func TestBadgerReversePrefixIterator(t *testing.T) {
	tmp := t.TempDir()
	dbInterface, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: tmp})
	require.NoError(t, err)

	bdb := dbInterface.(*badgerdb.BadgerDB)
	require.NotNil(t, bdb)

	_, values := fillBadgerValues(t, bdb)

	// Test reverse prefix iteration
	it := badgerdb.NewReversePrefixIterator(bdb, []byte("pre_"))
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

	defer it.Release()
	defer bdb.Close()
}

// TestBadgerReverseIteratorOrder verifies reverse order
func TestBadgerReverseIteratorOrder(t *testing.T) {
	tmp := t.TempDir()
	dbInterface, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: tmp})
	require.NoError(t, err)

	bdb := dbInterface.(*badgerdb.BadgerDB)
	require.NotNil(t, bdb)

	// Insert keys in predictable order
	keys := [][]byte{
		[]byte("key_01"),
		[]byte("key_02"),
		[]byte("key_03"),
		[]byte("key_04"),
		[]byte("key_05"),
	}
	for _, key := range keys {
		err := bdb.Put(t.Context(), key, []byte("value"))
		require.NoError(t, err)
	}

	// Get forward order
	forwardKeys := make([][]byte, 0)
	it := bdb.Scan([]byte("key_"))
	for it.Next() {
		forwardKeys = append(forwardKeys, it.Key())
	}
	it.Release()

	// Get reverse order
	reverseKeys := make([][]byte, 0)
	rit := badgerdb.NewReversePrefixIterator(bdb, []byte("key_"))
	for rit.Next() {
		reverseKeys = append(reverseKeys, rit.Key())
	}
	rit.Release()
	// Verify they're in opposite order
	require.Equal(t, len(forwardKeys), len(reverseKeys), "Should have same count")
	for i := 0; i < len(forwardKeys); i++ {
		require.Equal(t, forwardKeys[i], reverseKeys[len(reverseKeys)-1-i], "Keys should be in reverse order")
	}

	defer bdb.Close()
}
