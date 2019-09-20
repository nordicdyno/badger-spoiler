package spoiler_tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/y"
)

type heavyDataDir struct {
	sstFiles  []string
	vlogFiles []string
	manifest  string
}

func readHeavyDir(dbDir string) heavyDataDir {
	infos, err := ioutil.ReadDir(dbDir)
	if err != nil {
		panic(fmt.Sprintf("failed to read directory: %v", err.Error()))
	}
	var hd heavyDataDir
	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		info.Size()
		fInfo := filepath.Join(dbDir, info.Name())

		if info.Name() == "MANIFEST" {
			hd.manifest = fInfo
		} else if strings.HasSuffix(info.Name(), ".sst") {
			hd.sstFiles = append(hd.sstFiles, fInfo)
		} else if strings.HasSuffix(info.Name(), ".vlog") {
			hd.vlogFiles = append(hd.vlogFiles, fInfo)
		}
	}
	return hd
}

type KV struct {
	K []byte
	V []byte
}

func (kv KV) String() string {
	return fmt.Sprintf("%08b: %08b", kv.K, kv.V)
}

func spoilFileBits(path string, count int) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	size := fInfo.Size()
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	defer file.Close()
	if err != nil {
		return err
	}

	// TODO: improve algo, don't change same offsets
	fmt.Println("Spoil", path)
	for i := 0; i < count; i++ {
		offset := rand.Intn(int(size))

		orig := make([]byte, 1)
		n, err := file.ReadAt(orig, int64(offset))
		if err != nil {
			return err
		}
		fmt.Printf("read by offset  %v, %v bytes: % 02x\n", offset, n, orig)
		// fmt.Printf("\n", orig)

		bitOffset := rand.Intn(7)
		changed := flipBit(orig[0], bitOffset)

		n, err = file.WriteAt([]byte{changed}, int64(offset)) // Write at 0 beginning
		if err != nil {
			return err
		}
		fmt.Printf("write by offset %v, %v bytes: % 02x\n", offset, n, []byte{changed})
		// fmt.Printf("% 02x\n", []byte{changed})
	}

	return nil
}

func flipBit(b byte, offset int) byte {
	mask := byte(1 << uint8(offset))
	return b ^ mask
}

// func spoilHeavyFiles(dbDir string) {
// 	// TODO: implement spoiling
// 	infos, err := ioutil.ReadDir(dbDir)
// 	if err != nil {
// 		failF("failed to read directory: %v", err.Error())
// 	}
// 	var hd heavyDataDir
// 	for _, info := range infos {
// 		if info.IsDir() {
// 			continue
// 		}
// 		if info.Name() == "MANIFEST" {
// 			hd.manifest = filepath.Join(dbDir, "MANIFEST")
// 			continue
// 		}
// 		if strings.HasSuffix(info.Name(), ".sst") {
// 			hd.sstFiles = append(hd.sstFiles, filepath.Join(dbDir, info.Name()))
// 			continue
// 		}
// 		if strings.HasSuffix(info.Name(), ".vlog") {
// 			hd.vlogFiles = append(hd.vlogFiles, filepath.Join(dbDir, info.Name()))
// 			continue
// 		}
// 	}
// 	// fmt.Printf("heavyDataDir=> %#v\n", hd)
//
// 	var flags uint32
// 	fp, err := y.OpenExistingFile(hd.manifest, flags) // We explicitly sync in addChanges, outside the lock.
// 	if err != nil {
// 		failF("manifest open filed", err)
// 	}
// 	defer fp.Close()
//
// 	manifest, truncOffset, err := badger.ReplayManifestFile(fp)
// 	fmt.Printf("manifest: %#v\n", manifest)
// 	fmt.Printf("truncOffset: %#v\n", truncOffset)
// 	fmt.Printf("err: %#v\n", err)
// }

//
func openManifest(path string) error {
	var flags uint32
	flags |= y.ReadOnly

	fp, err := y.OpenExistingFile(path, flags) // We explicitly sync in addChanges, outside the lock.
	if err != nil {
		return err
	}
	defer fp.Close()

	manifest, truncOffset, err := badger.ReplayManifestFile(fp)
	fmt.Printf("manifest: %#v\n", manifest)
	fmt.Printf("truncOffset: %#v\n", truncOffset)
	fmt.Printf("err: %#v\n", err)
	return nil
}

// func openHeavyDir(dbDir string) error {
// 	opts := badger.DefaultOptions(dbDir)
// 	// do.ReadOnly = true
// 	db, err := badger.Open(opts)
// 	if err != nil {
// 		return err
// 	}
// 	tables := db.Tables(true)
// 	for i, t := range tables {
// 		fmt.Printf("%v table:\n", i)
// 		fmt.Println(strings.Repeat("-", 55))
// 		fmt.Printf("%#v\n\n", t)
// 	}
// 	return nil
// }
