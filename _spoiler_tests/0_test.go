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

var (
	iter = 0
	dataDir = filepath.Join("testdata", "gendb")
)


func Test_IterationsCounter(t *testing.T) {
	iter++
	t.Logf("ITERATION: %v", iter)
}

func Test_DBGen(t *testing.T) {
	killGenSig := os.Getenv("KILL_GEN_SIGNAL")

	err := os.RemoveAll(dataDir)
	require.NoErrorf(t, err, "cleanup dir with generated data: %v", dataDir)

	var genArgs []string
	if killGenSig == "" {
		genArgs = append(genArgs,
			"-keys", "1000",
			"-batchSize", "100",
			"-batchTimeout", "0",
			"-data-dir", dataDir,
		)
	}
	cmd := exec.Command(filepath.Join("..", "bin", "datagen"), genArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("RUN:", cmd.String())
	err = cmd.Start()
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
}

func Test_FlipBits(t *testing.T) {
	copyDir := filepath.Join("testdata", "gendb_copy")

	err := os.RemoveAll(copyDir)
	require.NoErrorf(t, err, "remove dir %v", copyDir)
	t.Logf("remove dir %v", copyDir)

	err = Copy(dataDir, copyDir)
	require.NoErrorf(t, err, "copy %v to %v", dataDir, copyDir)
	t.Logf("copy %v to %v", dataDir, copyDir)

	before, err := allKV(copyDir)
	require.NoErrorf(t, err, "%v data dir open", copyDir)

	// cleanup (because badger change opened files in non-readonly mode)
	// and make copy again
	err = os.RemoveAll(copyDir)
	require.NoErrorf(t, err, "remove dir %v", copyDir)
	t.Logf("remove dir %v", copyDir)

	err = Copy(dataDir, copyDir)
	require.NoErrorf(t, err, "copy %v to %v", dataDir, copyDir)
	t.Logf("copy %v to %v", dataDir, copyDir)

	dataDirInfo := readHeavyDir(copyDir)
	fmt.Printf("dataDirInfo => %#v\n", dataDirInfo)

	for _, fPath := range dataDirInfo.vlogFiles {
		t.Logf("spoil bits in %v", fPath)
		err := spoilFileBits(fPath, 1)
		require.NoErrorf(t, err, "spoil bits in %v", fPath)
	}

	after, err := allKV(copyDir)
	if err != nil {
		if _, ok := err.(BadgerOpenError); ok {
			t.Log("Got error on db open, all ok")
			return
		}
		require.NoErrorf(t, err, "unexpected error while reading keys")
	}
	t.Log("An error on spoiled DB is expected but got nil.")

	if len(before) != len(after) {
		t.Fatalf("originally had %v keys, but after spoiing vlog files got %v keys", len(before), len(after))
	}

	t.Run("keys_check", func(t *testing.T) {
		t.Log("keys_check", len(before))
		for i, kv := range before {
			require.Equalf(t, kv.K, after[i].K, "Check keys equality")
		}
	})
	t.Run("values_check", func(t *testing.T) {
		t.Log("keys_check", len(before))
		for i, kv := range before {
			require.Equalf(t, kv.V, after[i].V, "Check value of key=%x", kv.K)
		}
	})
}
