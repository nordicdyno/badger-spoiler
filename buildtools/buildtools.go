package buildtools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var BadgerDefaultVersion = "v1.6.0"

func InitBadgerModule(name string, dir string, badgerVersion string) error {
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
	var cmdBadgerDep *exec.Cmd
	if badgerVersion == "master" {
		badgerCheckoutDir, _ := filepath.Abs("_badger_src")
		if err := gitSources(badgerCheckoutDir, "https://" + badgerRepo, badgerVersion); err != nil {
			return err
		}

		cmdBadgerDep = exec.Command(
			"go", "mod", "edit",
			"-replace", badgerRepo+"="+badgerCheckoutDir,
		)
	} else {
		cmdBadgerDep = exec.Command(
			"go", "mod", "edit",
			"-require", badgerRepo+"@"+badgerVersion,
		)
	}
	cmdBadgerDep.Dir = dir
	cmdBadgerDep.Stdout = os.Stdout
	cmdBadgerDep.Stderr = os.Stderr

	if err := cmdBadgerDep.Run(); err != nil {
		return err
	}
	return nil
}

func gitSources(path string, repo string, revision string) error {
	if !dirExists(path) {
		if err := gitCommand(path, "clone", repo, path); err != nil {
			return err
		}
	}
	if err := gitCommand(path, "fetch", "--all", "--prune"); err != nil {
		return err
	}
	if err := gitCommand(path, "checkout", revision); err != nil {
		return err
	}
	if err := gitCommand(path, "pull"); err != nil {
		return err
	}

	return nil
}

func gitCommand(path string, args ...string) error {
	checkoutCmd := exec.Command("git", args...)
	if args[0] != "clone" {
		checkoutCmd.Dir = path
	}
	checkoutCmd.Stdout = os.Stdout
	checkoutCmd.Stderr = os.Stderr
	fmt.Println("RUN", checkoutCmd)
	if err := checkoutCmd.Run(); err != nil {
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

func dirExists(path string) bool {
	_, err := os.Stat(path);
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		panic(err)
	}
	return true
}
