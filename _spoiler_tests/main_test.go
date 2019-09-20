package spoiler_tests

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	e := m.Run()
	if e != 0 {
		fmt.Printf("Tests failed on %v iteration\n", iter)
	}
	os.Exit(e)
}
