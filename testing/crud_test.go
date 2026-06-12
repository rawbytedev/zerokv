package tests

import (
	"fmt"
	"testing"

	"github.com/rawbytedev/zerokv/testing/fixture"
	"github.com/stretchr/testify/require"
)

type test struct {
	name string
	fn   func(t *testing.T, name string)
}

func TestZeroKvImplementation(t *testing.T) {
	dbs := fixture.AllDBs
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
	db := fixture.SetupDB(t, name)
	defer db.Close()
	keys := make([][]byte, 10)
	values := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = fixture.RandomBytes(16)
		values[i] = fixture.RandomBytes(32)
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
}

// TestGetNonExistentKey tests retrieval of a non-existent key.
func testGetNonExistentKey(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	defer db.Close()
	nonExistentKey := fixture.RandomBytes(16)
	_, err := db.Get(t.Context(), nonExistentKey)
	require.Error(t, err, "Expected error when getting non-existent key")
}

// TestOverwriteKey tests overwriting an existing key.
func testOverwriteKey(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	defer db.Close()
	key := fixture.RandomBytes(16)
	value1 := fixture.RandomBytes(32)
	value2 := fixture.RandomBytes(32)
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

}

// TestClose tests closing the DB instance.
func testClose(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	err := db.Close()
	require.NoError(t, err, "Error closing DB")
}
