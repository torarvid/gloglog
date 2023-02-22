package testutil

import (
	"os"
	"testing"
)

func TestTempEnv(t *testing.T) {
	os.Setenv("FOO", "baz")

	func() {
		defer TempEnv("FOO", "bar")()
		if os.Getenv("FOO") != "bar" {
			t.Error("Expected FOO=bar")
		}
	}()
	if os.Getenv("FOO") != "baz" {
		t.Error("Expected FOO=baz")
	}
}
