package main

import (
	"context"

	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/pebbledb"
)

func multi_main() {
	/* Swapping kv database is simple and easy */

	data_db, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/temp"})
	idx_db, _ := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp"})

	defer data_db.Close()
	defer idx_db.Close()
	key := []byte("hello")
	/*
		It's easier to focus on logic, code is clear, and readable
		data_db store data while idx_db store the index to key
		and they both make use of Put() to store
	*/
	data_db.Put(context.Background(), key, []byte("world"))
	idx_db.Put(context.Background(), []byte{1}, key)
}
