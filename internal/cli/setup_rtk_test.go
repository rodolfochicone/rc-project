package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/setup"
	"github.com/spf13/cobra"
)

// newRTKTestState returns a setup state whose rtk dependencies are stubbed so
// no real PATH lookup or installer ever runs during unit tests.
func newRTKTestState() *setupCommandState {
	state := newSetupCommandState()
	state.goos = "linux"
	state.lookPath = func(string) (string, error) { return "", errors.New("not found") }
	state.confirmToolInstall = func(string, string) (bool, error) { return false, nil }
	state.runToolInstall = func(context.Context, setup.InstallCommand, io.Writer) error { return nil }
	return state
}

func runEnsureRTK(t *testing.T, state *setupCommandState) string {
	t.Helper()
	cmd := &cobra.Command{Use: "setup"}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := state.ensureRTK(context.Background(), cmd); err != nil {
		t.Fatalf("ensureRTK: %v\noutput:\n%s", err, out.String())
	}
	return out.String()
}

func TestEnsureRTKReportsDetectedBinary(t *testing.T) {
	t.Parallel()

	state := newRTKTestState()
	state.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		return setup.ToolStatus{Installed: true, Version: "rtk 1.2.3"}, nil
	}
	installCalled := false
	state.runToolInstall = func(context.Context, setup.InstallCommand, io.Writer) error {
		installCalled = true
		return nil
	}

	out := runEnsureRTK(t, state)
	if !strings.Contains(out, "rtk 1.2.3") {
		t.Fatalf("expected detected version in output, got:\n%s", out)
	}
	if installCalled {
		t.Fatal("installer must not run when rtk is already present")
	}
}

func TestEnsureRTKInteractiveConfirmRunsInstaller(t *testing.T) {
	t.Parallel()

	state := newRTKTestState()
	state.isInteractive = func() bool { return true }
	state.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		return setup.ToolStatus{Installed: false}, nil
	}
	state.confirmToolInstall = func(string, string) (bool, error) { return true, nil }

	var ranCommand setup.InstallCommand
	state.runToolInstall = func(_ context.Context, cmd setup.InstallCommand, _ io.Writer) error {
		ranCommand = cmd
		return nil
	}

	out := runEnsureRTK(t, state)
	if ranCommand.Name == "" {
		t.Fatal("expected installer to run after confirmation")
	}
	if !strings.Contains(out, "rtk installed") {
		t.Fatalf("expected success message, got:\n%s", out)
	}
	if !strings.Contains(out, "rtk init -g") {
		t.Fatalf("expected hook activation hint, got:\n%s", out)
	}
}

func TestEnsureRTKInteractiveDeclineSkipsInstaller(t *testing.T) {
	t.Parallel()

	state := newRTKTestState()
	state.isInteractive = func() bool { return true }
	state.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		return setup.ToolStatus{Installed: false}, nil
	}
	state.confirmToolInstall = func(string, string) (bool, error) { return false, nil }
	installCalled := false
	state.runToolInstall = func(context.Context, setup.InstallCommand, io.Writer) error {
		installCalled = true
		return nil
	}

	out := runEnsureRTK(t, state)
	if installCalled {
		t.Fatal("installer must not run when the user declines")
	}
	if !strings.Contains(out, "run later") {
		t.Fatalf("expected skip guidance, got:\n%s", out)
	}
}

func TestEnsureRTKYesModeOnlyInstructs(t *testing.T) {
	t.Parallel()

	state := newRTKTestState()
	state.yes = true
	state.isInteractive = func() bool { return true }
	state.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		return setup.ToolStatus{Installed: false}, nil
	}
	confirmCalled := false
	state.confirmToolInstall = func(string, string) (bool, error) {
		confirmCalled = true
		return true, nil
	}
	installCalled := false
	state.runToolInstall = func(context.Context, setup.InstallCommand, io.Writer) error {
		installCalled = true
		return nil
	}

	out := runEnsureRTK(t, state)
	if confirmCalled {
		t.Fatal("--yes must not prompt for rtk install")
	}
	if installCalled {
		t.Fatal("--yes must not run a network installer unattended")
	}
	if !strings.Contains(out, "curl -fsSL") {
		t.Fatalf("expected install command guidance, got:\n%s", out)
	}
}
