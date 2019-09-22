// +build !badger_v1.5

package spoiler_tests

import (
	"log"

	"github.com/dgraph-io/badger"
)

func allKV(dbDir string) ([]KV, error) {
	do := badger.DefaultOptions(dbDir)
	// do.ReadOnly = true
	db, err := badger.Open(do)
	if err != nil {
		return nil, BadgerOpenError{err}
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Printf("ERROR ON BADGER CLOSING %v", err)
		}
	}()

	var all []KV
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		// opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			var b []byte
			item := it.Item()
			// k :=
			v, err := item.ValueCopy(b)
			if err != nil {
				return err
			}
			all = append(all, KV{K: item.KeyCopy(b), V: v})
		}
		return nil
	})
	return all, err
}
