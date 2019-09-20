// +build badger_v1.5

package main

import (
	"crypto/rand"
	"time"

	"github.com/dgraph-io/badger"
)

func (b *batch) update() error {
	opts := badger.DefaultOptions
	opts.Dir = b.dbDir
	opts.ValueDir = b.dbDir
	// do.ReadOnly = true
	db, err := badger.Open(opts)
	if err != nil {
		failf("failed to open db %v: %v", b.dbDir, err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			failf("ERROR ON BADGER CLOSING %v", err)
		}
	}()

	batchSleepDuration := time.Duration(int64(time.Millisecond) * int64(b.batchTimeout))
	for i := 0; i < *keysCount; {
		// j := 0
		time.Sleep(batchSleepDuration)
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
				if err = txn.Set(k, v); err != nil {
					return err
				}
				i++
			}
			return nil
		})
		if err != nil {
			failf("write to db failed: %v", err)
		}
	}
	return nil
}
