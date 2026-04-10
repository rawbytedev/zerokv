package tests

import (
	"bytes"
	"fmt"
	"testing"

	zerokv "github.com/rawbytedev/zerokv/core"
	"github.com/rawbytedev/zerokv/testing/fixture"
	"github.com/stretchr/testify/require"
)

func FillValues(t *testing.T, db zerokv.Core) ([][]byte, [][]byte) {
	keys := make([][]byte, 10)
	values := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = fixture.RandomBytes(16)
		values[i] = fixture.RandomBytes(32)
		pref_key := make([]byte, 0)
		pref_key = append(pref_key, []byte("pre_")...)
		pref_key = append(pref_key, keys[i]...)
		err := db.Put(t.Context(), pref_key, values[i])
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}
	return keys, values
}

func TestZeroKvIterator(t *testing.T) {
	dbs := []string{"pebbledb", "badgerdb"}
	list_test := []test{
		{
			name: "TestIterateValue",
			fn: func(t *testing.T, name string) {
				testIterateValues(t, name)
			}}, {
			name: "testIterateHasKey",
			fn: func(t *testing.T, name string) {
				testIterateHasKey(t, name)
			},
		}, {
			name: "testIterateEmptyDatabase",
			fn: func(t *testing.T, name string) {
				testIterateEmptyDatabase(t, name)
			},
		}, {
			name: "testIterateSingleKey",
			fn: func(t *testing.T, name string) {
				testIterateSingleKey(t, name)
			},
		}, {
			name: "testIteratePrefixNotFound",
			fn: func(t *testing.T, name string) {
				testIteratePrefixNotFound(t, name)
			},
		}, {
			name: "testIterateVerifyValues",
			fn: func(t *testing.T, name string) {
				testIterateVerifyValues(t, name)
			},
		}, {
			name: "testIteratorErrorHandling",
			fn: func(t *testing.T, name string) {
				testIteratorErrorHandling(t, name)
			},
		}, {
			name: "testIterateKeysWithSpecialCharacters",
			fn: func(t *testing.T, name string) {
				testIterateKeysWithSpecialCharacters(t, name)
			},
		},
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

func testIterateValues(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	_, _ = FillValues(t, db)
	it := db.Scan([]byte("pre_"))
	for range 10 {
		require.True(t, it.Next())
		require.Equal(t, it.Key()[:len("pre_")], []byte("pre_"))
	}
	require.False(t, it.Next())
	defer db.Close()
	defer it.Release()
}

func testIterateHasKey(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	keys, _ := FillValues(t, db)
	it := db.Scan([]byte("pre_"))
	for range 10 {
		require.True(t, it.Next())
		cond := false
		key := it.Key()[len("pre_"):]
		for idx := range 10 {
			if bytes.Equal(keys[idx], key) {
				cond = true
				break
			}
		}
		require.True(t, cond)
	}
	require.False(t, it.Next())
	defer db.Close()
	defer it.Release()
}

// testIterateEmptyDatabase tests iteration over an empty database
func testIterateEmptyDatabase(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	it := db.Scan([]byte("pre_"))
	require.False(t, it.Next(), "Expected Next() to return false on empty database")
	require.Nil(t, it.Key(), "Expected Key() to return nil on empty database")
	require.Nil(t, it.Value(), "Expected Value() to return nil on empty database")
	require.NoError(t, it.Error(), "Expected no error on empty database iteration")
	defer db.Close()
	defer it.Release()
}

// testIterateSingleKey tests iteration with only one matching key
func testIterateSingleKey(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	key := []byte("pre_single")
	value := fixture.RandomBytes(32)
	err := db.Put(t.Context(), key, value)
	require.NoError(t, err)

	it := db.Scan([]byte("pre_"))
	require.True(t, it.Next(), "Expected Next() to return true for single key")
	require.Equal(t, key, it.Key(), "Key should match inserted key")
	require.Equal(t, value, it.Value(), "Value should match inserted value")
	require.False(t, it.Next(), "Expected Next() to return false after single key")

	defer db.Close()
	defer it.Release()
}

// testIteratePrefixNotFound tests iteration when prefix has no matches
func testIteratePrefixNotFound(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	// Insert keys with different prefix
	err := db.Put(t.Context(), []byte("other_key1"), fixture.RandomBytes(32))
	require.NoError(t, err)
	err = db.Put(t.Context(), []byte("other_key2"), fixture.RandomBytes(32))
	require.NoError(t, err)

	// Try to iterate with non-matching prefix
	it := db.Scan([]byte("pre_"))
	require.False(t, it.Next(), "Expected Next() to return false for non-matching prefix")

	defer db.Close()
	defer it.Release()
}

// testIterateVerifyValues tests that values from iterator match inserted values
func testIterateVerifyValues(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	_, values := FillValues(t, db)

	it := db.Scan([]byte("pre_"))
	count := 0
	for it.Next() {
		count++
		fullKey := it.Key()
		require.True(t, len(fullKey) > len("pre_"), "Key should contain prefix")
		require.Equal(t, fullKey[:len("pre_")], []byte("pre_"), "Key should start with prefix")

		// Verify value is not nil and has expected length
		retrievedValue := it.Value()
		require.NotNil(t, retrievedValue, "Value should not be nil")
		require.Equal(t, len(retrievedValue), 32, "Value should match inserted size")

		// Verify value matches one of the inserted values
		found := false
		for i := 0; i < len(values); i++ {
			if bytes.Equal(values[i], retrievedValue) {
				found = true
				break
			}
		}
		require.True(t, found, "Retrieved value should match one of inserted values")
	}
	require.Equal(t, 10, count, "Should iterate exactly 10 times")
	require.NoError(t, it.Error(), "Iterator should have no errors")

	defer db.Close()
	defer it.Release()
}

// testIteratorErrorHandling tests the Error() method
func testIteratorErrorHandling(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)
	_, values := FillValues(t, db)

	it := db.Scan([]byte("pre_"))
	for it.Next() {
		// Accessing value should work without errors
		retrievedValue := it.Value()
		require.NotNil(t, retrievedValue, "Value should not be nil")

		// Error should be nil for successful value retrieval
		iterErr := it.Error()
		require.NoError(t, iterErr, "Iterator should have no errors during iteration")

		// Verify the value is in our expected set
		found := false
		for i := 0; i < len(values); i++ {
			if bytes.Equal(values[i], retrievedValue) {
				found = true
				break
			}
		}
		require.True(t, found, "Retrieved value should be from inserted values")
	}

	defer db.Close()
	defer it.Release()
}

// testIterateKeysWithSpecialCharacters tests iteration with keys containing special characters
func testIterateKeysWithSpecialCharacters(t *testing.T, name string) {
	db := fixture.SetupDB(t, name)

	// Insert keys with special characters in prefix
	keys := [][]byte{
		[]byte("pre_\x00key1"),   // null byte
		[]byte("pre_\xFFkey2"),   // 0xFF byte
		[]byte("pre_\x01key3"),   // control character
		[]byte("pre_key\x00end"), // null in middle
	}

	for i, key := range keys {
		err := db.Put(t.Context(), key, []byte{byte(i)})
		require.NoError(t, err)
	}

	// Iterate and verify all special char keys are found
	it := db.Scan([]byte("pre_"))
	foundCount := 0
	for it.Next() {
		foundCount++
		fullKey := it.Key()
		require.True(t, bytes.HasPrefix(fullKey, []byte("pre_")), "Key should have correct prefix")

		// Verify key is one of our special char keys
		found := false
		for _, expectedKey := range keys {
			if bytes.Equal(fullKey, expectedKey) {
				found = true
				break
			}
		}
		require.True(t, found, "Found key should be in expected set")
	}

	require.Equal(t, 4, foundCount, "Should find all 4 keys with special characters")
	require.NoError(t, it.Error(), "Iterator should have no errors")

	defer db.Close()
	defer it.Release()
}
