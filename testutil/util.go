package testutil

import (
	"os"
	"testing"
)

func TempEnv(key, value string) (reset func()) {
	old := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		defer os.Setenv(key, old)
	}
}

func AssertEq[T comparable](t *testing.T, expected T, actual T) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
