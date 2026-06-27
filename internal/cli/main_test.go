package cli

import (
	"fmt"
	"os"
	"testing"
)

// TestMain isolates the CLI test suite from the developer's real ~/.rc/config.toml
// by pointing HOME at an empty temporary directory. Without this, every test that
// loads the global workspace config — directly via os.UserHomeDir, or via a
// subprocess launched with os.Environ() — depends on machine-local configuration,
// so a stale or unknown field in a developer's config fails otherwise-unrelated
// tests. Tests that need a specific HOME still override it via t.Setenv, which
// shadows and restores the value set here.
//
// The validate-tasks binary is *built* with the real HOME (see buildCLITestCommandEnv,
// which uses originalCLIHome captured at package init) so `go build` keeps using the
// real module cache; only the runtime HOME is isolated here.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "rc-cli-test-home-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cli test setup: create temp HOME: %v\n", err)
		os.Exit(1)
	}
	if err := os.Setenv("HOME", home); err != nil {
		_ = os.RemoveAll(home)
		fmt.Fprintf(os.Stderr, "cli test setup: set HOME: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = os.RemoveAll(home)
	os.Exit(code)
}
