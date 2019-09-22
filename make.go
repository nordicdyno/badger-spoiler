package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nordicdyno/badger-spoiler/buildtools"
)

var badgerVersion string

func main() {
	passArgs := argsParser()
	badgerBuildTag := buildtools.BadgerBuildTag(badgerVersion)
	fmt.Println("badgerVersion:", badgerVersion)
	fmt.Println("badgerBuildTag:", badgerBuildTag)
	fmt.Println("passArgs:", passArgs)

	// compile data generator
	buildModName := "datagen"
	buildDir := "_" + buildModName
	err := buildtools.InitModule(buildModName, buildDir, badgerVersion)
	if err != nil {
		log.Fatalf("datagen init module failed: %v", err)
	}

	cmdBuildGen := exec.Command("go", "build", "-v",
		"-tags", badgerBuildTag,
		"-o", filepath.Join("..", "bin", "datagen"),
		".")
	cmdBuildGen.Dir = buildDir
	cmdBuildGen.Stdout = os.Stdout
	cmdBuildGen.Stderr = os.Stderr

	fmt.Println("BUILD DATA GEN:", cmdBuildGen.String())
	if err = cmdBuildGen.Run(); err != nil {
		log.Fatalf("datagen binary compilation failed: %v", err)
	}

	testModName := "spoiler_tests"
	testDir := "_" + testModName
	err = buildtools.InitModule(testModName, testDir, badgerVersion)
	if err != nil {
		log.Fatalf("Failed init module %v in directory %v: %v\n", testDir, testDir, err)
	}

	// run tests
	args := []string{"test"}
	args = append(args, "-tags", badgerBuildTag)
	for _, arg := range passArgs {
		args = append(args, arg)
	}
	args = append(args, ".")
	cmdTest := exec.Command("go", args...)
	cmdTest.Dir = testDir
	cmdTest.Stdout = os.Stdout
	cmdTest.Stderr = os.Stderr

	fmt.Println("RUN TEST:", cmdTest.String())
	if err := cmdTest.Run(); err != nil {
		log.Fatalf("Failed run tests in directory %v: %v\n", testDir, err)
	}
}

func argsParser() []string {
	var pass bool
	var args []string
	var passArgs []string
	for _, arg := range os.Args[1:] {
		if arg == "--" {
			pass = true
			continue
		}
		if pass {
			passArgs = append(passArgs, arg)
			continue
		}
		args = append(args, arg)
	}

	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&badgerVersion, "v", buildtools.BadgerDefaultVersion, "badger version")

	err := f.Parse(args)
	if err != nil {
		panic(err)
	}
	return passArgs
}
