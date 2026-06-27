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

// stubInstaller rewires a tool installer so no real PATH lookup, TTY prompt, or
// installer runs during unit tests. Detection reports the tool missing by
// default; individual tests override detectTool/runInstall as needed.
func stubInstaller(i *toolInstaller) {
	i.goos = "linux"
	i.isInteractive = func() bool { return false }
	i.lookPath = func(string) (string, error) { return "", errors.New("not found") }
	i.confirm = func(string, string) (bool, error) { return false, nil }
	i.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		return setup.ToolStatus{Installed: false}, nil
	}
	i.runInstall = func(context.Context, setup.InstallCommand, io.Writer) error { return nil }
}

func newInstallTestState() *installCommandState {
	state := newInstallCommandState()
	for _, r := range state.resources {
		stubInstaller(r.installer)
	}
	return state
}

func (s *installCommandState) resource(t *testing.T, flag string) *installResource {
	t.Helper()
	for _, r := range s.resources {
		if r.flag == flag {
			return r
		}
	}
	t.Fatalf("resource %q not registered", flag)
	return nil
}

func runInstall(t *testing.T, state *installCommandState) (string, error) {
	t.Helper()
	cmd := &cobra.Command{Use: "install"}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := state.run(cmd, nil)
	return out.String(), err
}

func TestInstallWithoutResourceListsResources(t *testing.T) {
	t.Parallel()

	state := newInstallTestState()
	detectCalled := false
	for _, r := range state.resources {
		r.installer.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
			detectCalled = true
			return setup.ToolStatus{}, nil
		}
	}

	out, err := runInstall(t, state)
	if err != nil {
		t.Fatalf("install with no resource must not error, got: %v", err)
	}
	if detectCalled {
		t.Fatal("no detection should run when no resource is selected")
	}
	for _, want := range []string{"Installable resources", "--rtk", "--headroom"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected listing to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInstallRTKReportsDetectedBinary(t *testing.T) {
	t.Parallel()

	state := newInstallTestState()
	rtk := state.resource(t, "rtk")
	rtk.selected = true
	rtk.installer.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		return setup.ToolStatus{Installed: true, Version: "rtk 1.2.3"}, nil
	}
	installCalled := false
	rtk.installer.runInstall = func(context.Context, setup.InstallCommand, io.Writer) error {
		installCalled = true
		return nil
	}

	out, err := runInstall(t, state)
	if err != nil {
		t.Fatalf("install --rtk: %v", err)
	}
	if !strings.Contains(out, "rtk 1.2.3") {
		t.Fatalf("expected detected version in output, got:\n%s", out)
	}
	if installCalled {
		t.Fatal("installer must not run when rtk is already present")
	}
}

func TestInstallRTKYesRunsInstallerUnattended(t *testing.T) {
	t.Parallel()

	state := newInstallTestState()
	state.yes = true
	rtk := state.resource(t, "rtk")
	rtk.selected = true
	// Non-interactive on purpose: --yes must still install unattended, unlike setup.
	var ranCommand setup.InstallCommand
	rtk.installer.runInstall = func(_ context.Context, cmd setup.InstallCommand, _ io.Writer) error {
		ranCommand = cmd
		return nil
	}

	out, err := runInstall(t, state)
	if err != nil {
		t.Fatalf("install --rtk --yes: %v", err)
	}
	if ranCommand.Name != "sh" {
		t.Fatalf("expected the linux install script to run unattended, got %#v", ranCommand)
	}
	if !strings.Contains(out, "rtk installed") {
		t.Fatalf("expected success message, got:\n%s", out)
	}
}

func TestInstallHeadroomYesRunsPipInstaller(t *testing.T) {
	t.Parallel()

	state := newInstallTestState()
	state.yes = true
	hr := state.resource(t, "headroom")
	hr.selected = true
	// pip3 available, pipx/pip absent -> resolve to a pip3 install.
	hr.installer.lookPath = func(name string) (string, error) {
		if name == "pip3" {
			return "/usr/bin/pip3", nil
		}
		return "", errors.New("not found")
	}
	var ranCommand setup.InstallCommand
	hr.installer.runInstall = func(_ context.Context, cmd setup.InstallCommand, _ io.Writer) error {
		ranCommand = cmd
		return nil
	}

	out, err := runInstall(t, state)
	if err != nil {
		t.Fatalf("install --headroom --yes: %v", err)
	}
	if ranCommand.Name != "pip3" {
		t.Fatalf("expected pip3 installer, got %#v", ranCommand)
	}
	if !strings.Contains(strings.Join(ranCommand.Args, " "), "headroom-ai[all]") {
		t.Fatalf("expected headroom package in args, got %v", ranCommand.Args)
	}
	if !strings.Contains(out, "headroom installed") {
		t.Fatalf("expected success message, got:\n%s", out)
	}
}

func TestInstallGuideFlagShowsGuideWithoutInstalling(t *testing.T) {
	t.Parallel()

	state := newInstallTestState()
	state.guide = true
	rtk := state.resource(t, "rtk")
	rtk.selected = true
	detectCalled, installCalled := false, false
	rtk.installer.detectTool = func(context.Context, string) (setup.ToolStatus, error) {
		detectCalled = true
		return setup.ToolStatus{}, nil
	}
	rtk.installer.runInstall = func(context.Context, setup.InstallCommand, io.Writer) error {
		installCalled = true
		return nil
	}

	out, err := runInstall(t, state)
	if err != nil {
		t.Fatalf("install --rtk --guide: %v", err)
	}
	if detectCalled || installCalled {
		t.Fatal("--guide must neither detect nor install the tool")
	}
	for _, want := range []string{"Getting started", "rtk init -g", "https://github.com/rtk-ai/rtk"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected guide to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInstallRTKInteractiveDeclineSkipsInstaller(t *testing.T) {
	t.Parallel()

	state := newInstallTestState()
	rtk := state.resource(t, "rtk")
	rtk.selected = true
	rtk.installer.isInteractive = func() bool { return true }
	rtk.installer.confirm = func(string, string) (bool, error) { return false, nil }
	installCalled := false
	rtk.installer.runInstall = func(context.Context, setup.InstallCommand, io.Writer) error {
		installCalled = true
		return nil
	}

	out, err := runInstall(t, state)
	if err != nil {
		t.Fatalf("install --rtk: %v", err)
	}
	if installCalled {
		t.Fatal("installer must not run when the user declines")
	}
	if !strings.Contains(out, "run later") {
		t.Fatalf("expected skip guidance, got:\n%s", out)
	}
}
