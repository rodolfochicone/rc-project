package setup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveRTKInstall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		goos         string
		hasBrew      bool
		hasCargo     bool
		wantRunnable bool
		wantName     string
		wantInArgs   string
	}{
		{
			name:         "darwin prefers brew when available",
			goos:         "darwin",
			hasBrew:      true,
			wantRunnable: true,
			wantName:     "brew",
			wantInArgs:   "rtk",
		},
		{
			name:         "darwin falls back to install script without brew",
			goos:         "darwin",
			hasBrew:      false,
			wantRunnable: true,
			wantName:     "sh",
			wantInArgs:   rtkInstallScriptURL,
		},
		{
			name:         "linux uses install script",
			goos:         "linux",
			wantRunnable: true,
			wantName:     "sh",
			wantInArgs:   rtkInstallScriptURL,
		},
		{
			name:         "windows uses cargo when available",
			goos:         "windows",
			hasCargo:     true,
			wantRunnable: true,
			wantName:     "cargo",
			wantInArgs:   rtkRepoURL,
		},
		{
			name:         "windows without cargo is not runnable",
			goos:         "windows",
			hasCargo:     false,
			wantRunnable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ResolveRTKInstall(tt.goos, tt.hasBrew, tt.hasCargo)
			if got.Runnable != tt.wantRunnable {
				t.Fatalf("Runnable = %t, want %t (cmd=%#v)", got.Runnable, tt.wantRunnable, got)
			}
			if !tt.wantRunnable {
				// A non-runnable command must still guide the user manually.
				if strings.TrimSpace(got.Manual) == "" {
					t.Fatalf("expected non-runnable command to carry manual guidance, got %#v", got)
				}
				return
			}
			if got.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if strings.TrimSpace(got.Display) == "" {
				t.Fatal("expected runnable command to carry a display string")
			}
			if !strings.Contains(strings.Join(got.Args, " "), tt.wantInArgs) {
				t.Fatalf("args %v missing %q", got.Args, tt.wantInArgs)
			}
		})
	}
}

func TestDetectRTKReportsMissingBinary(t *testing.T) {
	// Not parallel: mutates PATH for the process via t.Setenv.
	t.Setenv("PATH", t.TempDir())

	status, err := DetectTool(context.Background(), RTKBinaryName)
	if err != nil {
		t.Fatalf("DetectTool: %v", err)
	}
	if status.Installed {
		t.Fatalf("expected rtk to be reported missing, got %#v", status)
	}
}

func TestDetectRTKReportsInstalledBinaryWithVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell-script binary is not executable on Windows")
	}

	dir := t.TempDir()
	writeFakeRTK(t, dir, "#!/bin/sh\necho 'rtk 9.9.9'\n")
	t.Setenv("PATH", dir)

	status, err := DetectTool(context.Background(), RTKBinaryName)
	if err != nil {
		t.Fatalf("DetectTool: %v", err)
	}
	if !status.Installed {
		t.Fatalf("expected rtk reported installed, got %#v", status)
	}
	if status.Version != "rtk 9.9.9" {
		t.Fatalf("Version = %q, want %q", status.Version, "rtk 9.9.9")
	}
}

func TestRunRTKInstall(t *testing.T) {
	t.Parallel()

	t.Run("rejects non-runnable command", func(t *testing.T) {
		t.Parallel()

		err := RunInstall(context.Background(), InstallCommand{Runnable: false}, &bytes.Buffer{})
		if err == nil {
			t.Fatal("expected error for non-runnable command")
		}
	})

	t.Run("streams output of a successful command", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("uses /bin/sh")
		}
		t.Parallel()

		var out bytes.Buffer
		cmd := InstallCommand{Name: "sh", Args: []string{"-c", "echo installed-ok"}, Runnable: true}
		if err := RunInstall(context.Background(), cmd, &out); err != nil {
			t.Fatalf("RunInstall: %v", err)
		}
		if !strings.Contains(out.String(), "installed-ok") {
			t.Fatalf("expected streamed output, got %q", out.String())
		}
	})

	t.Run("wraps failing command errors", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("uses /bin/sh")
		}
		t.Parallel()

		cmd := InstallCommand{Name: "sh", Args: []string{"-c", "exit 7"}, Display: "boom", Runnable: true}
		err := RunInstall(context.Background(), cmd, &bytes.Buffer{})
		if err == nil {
			t.Fatal("expected error for failing command")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("expected error to mention the command, got %v", err)
		}
	})
}

func writeFakeRTK(t *testing.T, dir, script string) {
	t.Helper()
	path := filepath.Join(dir, RTKBinaryName)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake rtk: %v", err)
	}
}
