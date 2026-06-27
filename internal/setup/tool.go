package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// ToolStatus reports whether an external CLI tool is available on the host.
type ToolStatus struct {
	// Installed is true when the binary was resolved on PATH.
	Installed bool
	// Path is the resolved executable path; empty when not installed.
	Path string
	// Version is the trimmed `<binary> --version` output; empty when not
	// installed or when the version probe failed.
	Version string
}

// InstallCommand describes the OS/environment-appropriate way to install a tool.
type InstallCommand struct {
	// Display is the human-readable command shown in prompts and guidance.
	Display string
	// Name and Args form the executable invocation when Runnable is true.
	Name string
	Args []string
	// Runnable reports whether rc can execute the command directly.
	Runnable bool
	// Manual carries fallback guidance shown when Runnable is false.
	Manual string
}

// DetectTool resolves binary on PATH and probes its version. A missing binary
// is reported as a non-error ToolStatus with Installed=false because an absent
// optional tool is an expected outcome, not a failure.
func DetectTool(ctx context.Context, binary string) (ToolStatus, error) {
	path, err := exec.LookPath(binary)
	if err != nil {
		// LookPath only fails when the binary is absent or not executable;
		// both mean the tool is simply unavailable for this run.
		return ToolStatus{}, nil
	}

	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		// The binary exists but the version probe failed; still report it
		// installed so callers do not offer a redundant install.
		return ToolStatus{Installed: true, Path: path}, nil
	}
	return ToolStatus{
		Installed: true,
		Path:      path,
		Version:   strings.TrimSpace(string(out)),
	}, nil
}

// RunInstall executes the resolved install command, streaming combined output
// to w. It returns an error when the command is not runnable on this platform
// or when execution fails.
func RunInstall(ctx context.Context, command InstallCommand, w io.Writer) error {
	if !command.Runnable {
		return errors.New("install command is not runnable on this platform")
	}
	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %q: %w", command.Display, err)
	}
	return nil
}
