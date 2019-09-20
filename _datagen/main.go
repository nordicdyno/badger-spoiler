package main

import (
	"flag"
	"log"
	"os"
)

var (
	keysCount    = flag.Int("keys", 10000, "how many keys generate")
	batchSize    = flag.Int("batchSize", 10, "batch size")
	batchTimeout = flag.Int("batchTimeout", 100, "timeout before batch in milliseconds")
	keySize      = flag.Int("keySize", 100, "key size in bytes")
	valueSize    = flag.Int("valueSize", 1000, "value size in bytes")

	dbDir = flag.String("data-dir", "testdata/gendb", "testdata")
)

type batch struct {
	dbDir        string
	batchTimeout int
	batchSize    int
	keysCount    int
	keySize      int
	valueSize    int
}

func main() {
	flag.Parse()

	if err := os.Mkdir(*dbDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create db %v: %v (%T)", *dbDir, err, err)
	}

	b := &batch{
		dbDir:        *dbDir,
		batchTimeout: *batchTimeout,
		batchSize:    *batchSize,
		keysCount:    *keysCount,
		keySize:      *keySize,
		valueSize:    *valueSize,
	}

	if err := b.update(); err != nil {
		log.Fatalf("BATCH ERROR: %v", err)
	}
}
