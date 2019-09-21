package spoiler_tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
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

		bitOffset := rand.Intn(7)
		changed := flipBit(orig[0], bitOffset)

		n, err = file.WriteAt([]byte{changed}, int64(offset)) // Write at 0 beginning
		if err != nil {
			return err
		}
		fmt.Printf("write by offset %v, %v bytes: % 02x\n", offset, n, []byte{changed})
	}

	return nil
}

func flipBit(b byte, offset int) byte {
	mask := byte(1 << uint8(offset))
	return b ^ mask
}
