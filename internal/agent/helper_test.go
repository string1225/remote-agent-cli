package agent

import (
	"io"
	"os"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("RA_TEST_HELPER") != "1" {
		return
	}
	_, err := io.Copy(os.Stdout, os.Stdin)
	if err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}
