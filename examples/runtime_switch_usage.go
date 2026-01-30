/*
Switch during runtimes to recover from errors or to balance load
*/
package main

import (
	"context"

	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/pebbledb"
)

func runtime_switching() {
	data_db, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/temp"})
	backup_db, _ := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp"})

	defer data_db.Close()
	defer backup_db.Close()
	key := []byte("hello")
	value := []byte("world")
	/* Populating both DB */
	data_db.Put(context.Background(), key, value)
	backup_db.Put(context.Background(), key, value)
	// Fetching from main DB
	retrieved, err := data_db.Get(context.Background(), key)
	if err != nil {
		// Fetching from Backup because main DB failed
		retrieved, err = backup_db.Get(context.Background(), key)
	}
	_ = retrieved

}
