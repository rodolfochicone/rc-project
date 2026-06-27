package cli

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateRepoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain name", input: "my-service", want: "my-service"},
		{name: "trims surrounding space", input: "  api  ", want: "api"},
		{name: "empty", input: "   ", wantErr: true},
		{name: "rejects owner slug", input: "rodolfochicone/api", wantErr: true},
		{name: "rejects path", input: "a/b", wantErr: true},
		{name: "rejects embedded space", input: "my service", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := validateRepoName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateRepoName(%q) = %q, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateRepoName(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("validateRepoName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

type ghCall struct {
	dir  string
	args []string
}

// fakeGH stands in for the gh CLI boundary so tests exercise rc init's real
// behavior (argument construction, error mapping) without spawning a process.
type fakeGH struct {
	authErr   error
	createOut []byte
	createErr error
	calls     []ghCall
}

func (f *fakeGH) run(_ context.Context, dir string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, ghCall{dir: dir, args: args})
	if len(args) > 0 && args[0] == "auth" {
		return nil, f.authErr
	}
	return f.createOut, f.createErr
}

func (f *fakeGH) createCall(t *testing.T) ghCall {
	t.Helper()
	for _, call := range f.calls {
		if len(call.args) > 0 && call.args[0] == "repo" {
			return call
		}
	}
	t.Fatalf("no `gh repo create` call recorded; calls=%v", f.calls)
	return ghCall{}
}

func newTestInitState(fake *fakeGH, lookPath func(string) (string, error)) *initCommandState {
	return &initCommandState{
		lookPath: lookPath,
		runGH:    fake.run,
		getwd:    func() (string, error) { return "/work", nil },
	}
}

func ghFound(string) (string, error) { return "/usr/bin/gh", nil }

func runInit(state *initCommandState, args ...string) (string, error) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := state.run(cmd, args)
	return out.String(), err
}

func TestInitRunSuccess(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{}
	state := newTestInitState(fake, ghFound)

	out, err := runInit(state, "my-service")
	if err != nil {
		t.Fatalf("run: unexpected error: %v", err)
	}

	create := fake.createCall(t)
	wantArgs := []string{
		"repo", "create", "rodolfochicone/my-service",
		"--template", "rodolfochicone/typescript-template",
		"--private", "--clone",
	}
	if !slices.Equal(create.args, wantArgs) {
		t.Fatalf("create args = %v, want %v", create.args, wantArgs)
	}
	if create.dir != "/work" {
		t.Fatalf("create dir = %q, want %q", create.dir, "/work")
	}
	if !strings.Contains(out, "rodolfochicone/my-service") || !strings.Contains(out, "./my-service") {
		t.Fatalf("output missing repo/clone path:\n%s", out)
	}
}

func TestInitRunPromptsWhenNameOmittedInteractive(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{}
	state := newTestInitState(fake, ghFound)
	state.isInteractive = func() bool { return true }
	state.promptName = func() (string, error) { return "prompted-service", nil }

	out, err := runInit(state) // no positional name
	if err != nil {
		t.Fatalf("run: unexpected error: %v", err)
	}

	create := fake.createCall(t)
	if got := create.args[2]; got != "rodolfochicone/prompted-service" {
		t.Fatalf("create target = %q, want rodolfochicone/prompted-service", got)
	}
	if !strings.Contains(out, "./prompted-service") {
		t.Fatalf("output missing prompted clone path:\n%s", out)
	}
}

func TestInitRunRequiresNameWhenNonInteractive(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{}
	state := newTestInitState(fake, ghFound)
	state.isInteractive = func() bool { return false }
	state.promptName = func() (string, error) {
		t.Fatal("prompt must not run in non-interactive mode")
		return "", nil
	}

	_, err := runInit(state) // no positional name
	if err == nil {
		t.Fatal("expected error when name is omitted in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "repository name is required") {
		t.Fatalf("error should ask for the name: %v", err)
	}
	if len(fake.calls) != 0 {
		t.Fatalf("expected no gh invocations without a name, got %v", fake.calls)
	}
}

func TestInitRunGitHubCLIMissing(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{}
	state := newTestInitState(fake, func(string) (string, error) { return "", exec.ErrNotFound })

	_, err := runInit(state, "my-service")
	if !errors.Is(err, errGitHubCLIMissing) {
		t.Fatalf("err = %v, want errGitHubCLIMissing", err)
	}
	if !strings.Contains(err.Error(), "gh auth login") {
		t.Fatalf("missing setup guidance in error:\n%v", err)
	}
	if len(fake.calls) != 0 {
		t.Fatalf("expected no gh invocations when gh is absent, got %v", fake.calls)
	}
}

func TestInitRunNotAuthenticated(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{authErr: errors.New("not logged in")}
	state := newTestInitState(fake, ghFound)

	_, err := runInit(state, "my-service")
	if !errors.Is(err, errGitHubNotAuthenticated) {
		t.Fatalf("err = %v, want errGitHubNotAuthenticated", err)
	}
	// The repo must not be created when authentication failed.
	if slices.ContainsFunc(fake.calls, func(c ghCall) bool {
		return len(c.args) > 0 && c.args[0] == "repo"
	}) {
		t.Fatalf("repo create attempted despite auth failure: %v", fake.calls)
	}
}

func TestInitRunOrgAccessDenied(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{
		createOut: []byte("HTTP 403: Resource not accessible by integration; you must be a member of rodolfochicone"),
		createErr: errors.New("exit status 1"),
	}
	state := newTestInitState(fake, ghFound)

	_, err := runInit(state, "my-service")
	if !errors.Is(err, errGitHubOrgAccess) {
		t.Fatalf("err = %v, want errGitHubOrgAccess", err)
	}
	if !strings.Contains(err.Error(), "read:org") {
		t.Fatalf("missing org-access guidance in error:\n%v", err)
	}
}

func TestInitRunGenericCreateError(t *testing.T) {
	t.Parallel()

	fake := &fakeGH{
		createOut: []byte("name already exists on this account"),
		createErr: errors.New("exit status 1"),
	}
	state := newTestInitState(fake, ghFound)

	_, err := runInit(state, "my-service")
	if err == nil {
		t.Fatal("expected error for failed create")
	}
	if errors.Is(err, errGitHubOrgAccess) {
		t.Fatalf("generic failure misclassified as org-access denial: %v", err)
	}
	if !strings.Contains(err.Error(), "name already exists") {
		t.Fatalf("error should surface gh output:\n%v", err)
	}
}
