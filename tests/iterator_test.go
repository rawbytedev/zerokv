package tests

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/helpers"
	"github.com/stretchr/testify/require"
)

func FillValues(t *testing.T, db zerokv.Core) ([][]byte, [][]byte) {
	keys := make([][]byte, 10)
	values := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = helpers.RandomBytes(16)
		values[i] = helpers.RandomBytes(32)
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
	db := helpers.SetupDB(t, name)
	_, _ = FillValues(t, db)
	it := db.Scan([]byte("pre_"))
	for range 10 {
		require.True(t, it.Next())
		require.Equal(t, it.Key()[:len("pre_")], []byte("pre_"))
	}
	defer db.Close()
	defer it.Release()
}

func testIterateHasKey(t *testing.T, name string) {
	db := helpers.SetupDB(t, name)
	keys, _ := FillValues(t, db)
	it := db.Scan([]byte("pre_"))
	for range 10 {
		it.Next()
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
	defer db.Close()
	defer it.Release()
}
