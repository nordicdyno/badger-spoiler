package spoiler_tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var iter = 0

func Test_IterationsCounter(t *testing.T) {
	iter++
	t.Logf("ITERATION: %v", iter)
}

func Test_DBGen(t *testing.T) {
	killGenSig := os.Getenv("KILL_GEN_SIGNAL")

	var genArgs []string
	if killGenSig == "" {
		genArgs = append(genArgs,
			"-keys", "1000",
			"-batchSize", "100",
			"-batchTimeout", "0",
		)
	}
	cmd := exec.Command("../bin/datagen", genArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("RUN:", cmd.String())
	err := cmd.Start()
	require.NoError(t, err)

	cmdResult := make(chan error)
	go func() {
		cmdResult <- cmd.Wait()
		fmt.Println("datagen done")
	}()

	killChan := make(<-chan time.Time)
	stopTimer := func() {}
	if killGenSig == "" {
		killChan = nil
	} else {
		killTimer := time.NewTimer(time.Millisecond * 200)
		killChan = killTimer.C
		stopTimer = func() { killTimer.Stop() }
	}

waitCommandLoop:
	for {
		select {
		case cmdErr := <-cmdResult:
			fmt.Printf("COMMAND Result: %v (%T)\n", cmdErr, cmdErr)
			if e, ok := cmdErr.(*exec.ExitError); ok {
				fmt.Printf("exit code=%v (%v)\n", e.ExitCode(), e.Error())
				t.Fatal("datagen failed")
			}
			break waitCommandLoop
		case <-killChan:
			stopTimer()
			pid := fmt.Sprintf("%v", cmd.Process.Pid)
			_, killErr := exec.Command("kill", "-"+killGenSig, pid).Output()
			if killErr != nil {
				t.Fatal("failed to kill process by pid", pid)
			}
		}
	}
	fmt.Println("Test_DBGen END")
}

func Test_FlipBits(t *testing.T) {
	fmt.Println("Test_FlipBits start")
	// origDirName := "_orig"
	origDirName := "gendb"

	copyDir := filepath.Join("testdata", "gendb_copy")

	err := os.RemoveAll(copyDir)
	require.NoErrorf(t, err, "remove dir %v", copyDir)
	t.Logf("remove dir %v", copyDir)

	origDir := filepath.Join("testdata", origDirName)
	err = Copy(origDir, copyDir)
	require.NoErrorf(t, err, "copy %v to %v", origDir, copyDir)
	t.Logf("copy %v to %v", origDir, copyDir)

	before, err := allKV(copyDir)
	require.NoErrorf(t, err, "%v data dir open", copyDir)

	// cleanup (because badger change opened files in non-readonly mode)
	// and make copy again
	err = os.RemoveAll(copyDir)
	require.NoErrorf(t, err, "remove dir %v", copyDir)
	t.Logf("remove dir %v", copyDir)

	err = Copy(origDir, copyDir)
	require.NoErrorf(t, err, "copy %v to %v", origDir, copyDir)
	t.Logf("copy %v to %v", origDir, copyDir)

	dataDirInfo := readHeavyDir(copyDir)
	fmt.Printf("dataDirInfo => %#v\n", dataDirInfo)

	for _, fPath := range dataDirInfo.vlogFiles {
		t.Logf("spoil bits in %v", fPath)
		err := spoilFileBits(fPath, 1)
		require.NoErrorf(t, err, "spoil bits in %v", fPath)
	}

	after1, err := allKV(copyDir)
	if err != nil {
		return
	}
	// t.Log("An error on spoiled DB is expected but got nil.")
	t.Error("An error on spoiled DB is expected but got nil.")

	// assert.Errorf(t, err, "%v data dir open", copyDir)
	if len(before) != len(after1) {
		// lostKeys
		t.Fatalf("originally had %v keys, but after spoiing vlog files got %v keys", len(before), len(after1))
	}

	// after first open database
	dataDirInfo = readHeavyDir(copyDir)
	fmt.Printf("dataDirInfo => %#v\n", dataDirInfo)

	after2, err := allKV(copyDir)
	require.NoErrorf(t, err, "%v data dir open", copyDir)
	if len(before) != len(after2) {
		// lostKeys
		t.Fatalf("was %v keys, but after spoil got %v keys", len(before), len(after2))
	}

	t.Run("keys_check", func(t *testing.T) {
		t.Log("keys_check", len(before))
		for i, kv := range before {
			require.Equalf(t, kv.K, after2[i].K, "Check keys equality")
		}
	})
	t.Run("values_check", func(t *testing.T) {
		t.Log("keys_check", len(before))
		for i, kv := range before {
			require.Equalf(t, kv.V, after2[i].V, "Check value of key=%x", kv.K)
		}
	})
}
