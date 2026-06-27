package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/rodolfochicone/rc-project/internal/setup"
)

// toolInstaller detects an external CLI tool and, when missing, surfaces and
// optionally runs the environment-appropriate installer. It powers both
// `rc install --<tool>` and the rtk step of `rc setup`. Dependencies are
// injectable so the flow is unit-testable without touching PATH or running an
// installer.
type toolInstaller struct {
	// binary is the executable resolved on PATH (e.g. "rtk", "headroom").
	binary string
	// label is the section header rendered in output (e.g. "RTK").
	label string
	// notFoundHint is the message shown when the binary is missing.
	notFoundHint string
	// guide is the "Getting started" tutorial printed after install/detection
	// and on demand via `rc install --<tool> --guide`.
	guide toolGuide
	// resolve returns the install command for the current environment. goos is
	// a runtime.GOOS value and onPath reports whether a helper binary (brew,
	// pipx, …) is available.
	resolve func(goos string, onPath func(name string) bool) setup.InstallCommand

	// yes skips the interactive confirmation. Combined with runUnattended it
	// decides whether the installer may run without a prompt.
	yes bool
	// runUnattended controls the --yes policy: setup only prints guidance
	// (false) while `rc install` actually runs the installer (true).
	runUnattended bool
	goos          string

	isInteractive func() bool
	detectTool    func(context.Context, string) (setup.ToolStatus, error)
	runInstall    func(context.Context, setup.InstallCommand, io.Writer) error
	confirm       func(label, display string) (bool, error)
	lookPath      func(string) (string, error)
}

// toolGuide is a short getting-started tutorial for an installed tool.
type toolGuide struct {
	// steps are example commands with a one-line description each.
	steps []toolGuideStep
	// docsURL points to the tool's official documentation.
	docsURL string
}

type toolGuideStep struct {
	command string
	desc    string
}

// newRTKInstaller returns an installer for the rtk runtime toolkit wired with
// production dependencies.
func newRTKInstaller() *toolInstaller {
	return &toolInstaller{
		binary:       setup.RTKBinaryName,
		label:        "RTK",
		notFoundHint: "rtk not found on PATH",
		guide: toolGuide{
			steps: []toolGuideStep{
				{command: "rtk init -g", desc: "enable the Claude Code hook (restart Claude Code afterward)"},
				{command: "rtk --help", desc: "list all available commands"},
			},
			docsURL: "https://github.com/rtk-ai/rtk",
		},
		resolve: func(goos string, onPath func(string) bool) setup.InstallCommand {
			return setup.ResolveRTKInstall(goos, onPath("brew"), onPath("cargo"))
		},
		goos:          runtime.GOOS,
		isInteractive: isInteractiveTerminal,
		detectTool:    setup.DetectTool,
		runInstall:    setup.RunInstall,
		confirm:       confirmToolInstall,
		lookPath:      exec.LookPath,
	}
}

// newHeadroomInstaller returns an installer for the headroom AI toolkit wired
// with production dependencies.
func newHeadroomInstaller() *toolInstaller {
	return &toolInstaller{
		binary:       setup.HeadroomBinaryName,
		label:        "Headroom",
		notFoundHint: "headroom not found on PATH",
		guide: toolGuide{
			steps: []toolGuideStep{
				{command: "headroom wrap claude", desc: "wrap a coding agent (Claude, Cursor, Aider, Copilot)"},
				{command: "headroom proxy --port 8787", desc: "start a local compression proxy for any client"},
				{command: "headroom perf", desc: "show real token savings"},
				{command: "headroom update", desc: "upgrade headroom to the latest version"},
			},
			docsURL: "https://github.com/headroomlabs-ai/headroom",
		},
		resolve: func(_ string, onPath func(string) bool) setup.InstallCommand {
			return setup.ResolveHeadroomInstall(onPath("pipx"), onPath("pip3"), onPath("pip"))
		},
		goos:          runtime.GOOS,
		isInteractive: isInteractiveTerminal,
		detectTool:    setup.DetectTool,
		runInstall:    setup.RunInstall,
		confirm:       confirmToolInstall,
		lookPath:      exec.LookPath,
	}
}

// ensure detects the tool and, when missing, surfaces the install command,
// running it only when policy permits (see shouldRun). It never modifies any
// downstream configuration; it only hints how to enable it.
func (t *toolInstaller) ensure(ctx context.Context, w io.Writer) error {
	status, err := t.detectTool(ctx, t.binary)
	if err != nil {
		return fmt.Errorf("detect %s: %w", t.binary, err)
	}

	if status.Installed {
		printToolDetected(w, t.label, t.binary, status)
		t.printGuide(w)
		return nil
	}

	install := t.resolve(t.goos, t.binaryOnPath)
	printToolMissing(w, t.label, t.notFoundHint, install)
	if !install.Runnable {
		return nil
	}

	confirmed, err := t.shouldRun(install)
	if err != nil {
		return err
	}
	if !confirmed {
		printToolSkipped(w, install)
		return nil
	}

	if err := t.runInstall(ctx, install, w); err != nil {
		return fmt.Errorf("install %s: %w", t.binary, err)
	}
	printToolInstalled(w, t.binary)
	t.printGuide(w)
	return nil
}

// printGuide renders the tool's getting-started tutorial. It is a no-op when
// the tool has no guide configured.
func (t *toolInstaller) printGuide(w io.Writer) {
	if len(t.guide.steps) == 0 && strings.TrimSpace(t.guide.docsURL) == "" {
		return
	}
	styles := newCLIChromeStyles()
	lipgloss.Fprintln(w, styles.sectionTitle.Render(t.label+" — Getting started"))

	width := 0
	for _, step := range t.guide.steps {
		if len(step.command) > width {
			width = len(step.command)
		}
	}
	for _, step := range t.guide.steps {
		lipgloss.Fprintf(
			w,
			"  %s  %s\n",
			styles.value.Render(padRight(step.command, width)),
			styles.path.Render(step.desc),
		)
	}
	if url := strings.TrimSpace(t.guide.docsURL); url != "" {
		lipgloss.Fprintf(w, "  %s  %s\n", styles.label.Render("Docs"), styles.value.Render(url))
	}
}

// shouldRun reports whether the installer may execute. With --yes the
// runUnattended policy decides; otherwise an interactive terminal and an
// explicit confirmation are required, and non-interactive sessions only
// receive guidance.
func (t *toolInstaller) shouldRun(install setup.InstallCommand) (bool, error) {
	if t.yes {
		return t.runUnattended, nil
	}
	if !t.isInteractive() {
		return false, nil
	}
	return t.confirm(t.label, install.Display)
}

func (t *toolInstaller) binaryOnPath(name string) bool {
	_, err := t.lookPath(name)
	return err == nil
}

func confirmToolInstall(label, display string) (bool, error) {
	confirmed := false
	field := huh.NewConfirm().
		Key("tool-install").
		Title(fmt.Sprintf("%s not found. Run the installer now?", label)).
		Description(display).
		Value(&confirmed)
	if err := runPromptField(field); err != nil {
		return false, fmt.Errorf("confirm %s install: %w", label, err)
	}
	return confirmed, nil
}

func printToolDetected(w io.Writer, label, binary string, status setup.ToolStatus) {
	styles := newCLIChromeStyles()
	lipgloss.Fprintln(w, styles.sectionTitle.Render(label))
	version := strings.TrimSpace(status.Version)
	if version == "" {
		version = binary + " detected"
	}
	lipgloss.Fprintf(w, "  %s  %s\n", styles.successIcon.Render("✓"), styles.value.Render(version))
}

func printToolMissing(w io.Writer, label, notFoundHint string, install setup.InstallCommand) {
	styles := newCLIChromeStyles()
	lipgloss.Fprintln(w, styles.sectionTitle.Render(label))
	lipgloss.Fprintf(w, "  %s  %s\n", styles.warn.Render("!"), styles.path.Render(notFoundHint))

	guidance := install.Display
	if !install.Runnable {
		guidance = install.Manual
	}
	lipgloss.Fprintf(w, "  %s  %s\n", styles.label.Render("Install"), styles.value.Render(guidance))
}

func printToolSkipped(w io.Writer, install setup.InstallCommand) {
	styles := newCLIChromeStyles()
	lipgloss.Fprintf(w, "  %s  %s\n", styles.label.Render("Skipped"), styles.path.Render("run later: "+install.Display))
}

func printToolInstalled(w io.Writer, binary string) {
	styles := newCLIChromeStyles()
	lipgloss.Fprintf(w, "  %s  %s\n", styles.successIcon.Render("✓"), styles.value.Render(binary+" installed"))
}
