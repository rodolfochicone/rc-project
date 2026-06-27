package test

import (
	"fmt"
	"os"
	"testing"
)

// TestMain isolates the public-API test suite from the developer's real
// ~/.rc/config.toml by pointing HOME at an empty temporary directory, so global
// configuration present on the machine cannot affect these tests.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "rc-public-api-test-home-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "public-api test setup: create temp HOME: %v\n", err)
		os.Exit(1)
	}
	if err := os.Setenv("HOME", home); err != nil {
		_ = os.RemoveAll(home)
		fmt.Fprintf(os.Stderr, "public-api test setup: set HOME: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = os.RemoveAll(home)
	os.Exit(code)
}
