package main

import (
	"bytes"
	"context"
	"fmt"

	// "github.com/rawbytedev/zerokv/pebbledb" // swapping db requires import the KV implementation
	"github.com/rawbytedev/zerokv/badgerdb"
)

func basic_main() {
	/*Swapping kv database is simple and easy*/
	db, _ := badgerdb.NewBadgerDB(badgerdb.Config{Dir: "/temp"})
	// db, - := pebbledb.NewPebbleDB(pebbledb.Config{Dir: "/tmp"})

	/*
		All other part of codes remain untouched as they don't need to be modified
		Save more than 70% time when changing kv databases
	*/
	defer db.Close()
	key := []byte("hello")
	db.Put(context.Background(), key, []byte("world"))
	value, err := db.Get(context.Background(), key)
	if err != nil {
		return
	}
	if bytes.Equal(value, []byte("world")) {
		fmt.Print("Value Retrieved Successfully")
	}
}
