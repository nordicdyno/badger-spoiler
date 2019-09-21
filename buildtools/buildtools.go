package buildtools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var BadgerDefaultVersion = "v1.6.0"

func InitModule(name string, dir string, badgerVersion string) error {
	goModFile := filepath.Join(dir, "go.mod")
	_, err := os.Stat(goModFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		// fmt.Println("err:", err, "os.IsNotExist(err)", os.IsNotExist(err))
		if err := os.Remove(goModFile); err != nil {
			return err
		}
	}

	cmdMod := exec.Command("go", "mod", "init", name)
	cmdMod.Dir = dir
	cmdMod.Stdout = os.Stdout
	cmdMod.Stderr = os.Stderr
	if err := cmdMod.Run(); err != nil {
		return err
	}

	// hint: go list -m -versions github.com/dgraph-io/badger
	if strings.HasPrefix(badgerVersion, "v2") {
		badgerVersion = badgerVersion + "+incompatible"
	}

	badgerRepo := "github.com/dgraph-io/badger"
	cmdEdit := exec.Command(
		"go", "mod", "edit",
		"-require", badgerRepo+"@"+badgerVersion,
	)
	cmdEdit.Dir = dir
	cmdEdit.Stdout = os.Stdout
	cmdEdit.Stderr = os.Stderr

	if err := cmdEdit.Run(); err != nil {
		return err
	}
	return nil
}

func BadgerBuildTag(version string) string {
	badgerBuildTag := version
	idx := strings.LastIndex(version, ".")
	if idx != -1 {
		badgerBuildTag = version[:idx]
	}
	return "badger_" + badgerBuildTag
}
