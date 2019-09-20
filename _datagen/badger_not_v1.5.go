// +build !badger_v1.5

package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger"
)

func (b *batch) update() error {
	opts := badger.DefaultOptions(b.dbDir)
	// do.ReadOnly = true
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open db %v: %v", b.dbDir, err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Println(fmt.Errorf("ERROR ON BADGER CLOSING %w", err))
		}
	}()

	batchSleepDuration := time.Duration(int64(time.Millisecond) * int64(b.batchTimeout))
	n := 0
	for i := 0; i < *keysCount; {
		n++
		time.Sleep(batchSleepDuration)
		// fmt.Printf("Load batch %v\n", n)
		err := db.Update(func(txn *badger.Txn) error {
			// batch
			var err error
			for j := 0; j < b.batchSize && i < b.keysCount; j++ {
				k := make([]byte, b.keySize)
				_, err = rand.Read(k)
				if err != nil {
					return err
				}
				v := make([]byte, b.valueSize)
				_, err = rand.Read(v)
				if err != nil {
					return err
				}
				e := badger.NewEntry(k, v)
				if err = txn.SetEntry(e); err != nil {
					return err
				}
				i++
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("write to db failed: %w", err)
		}
	}
	return nil
}
