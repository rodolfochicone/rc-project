package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

const (
	// initRepoOwner is the GitHub organization that always owns repositories
	// scaffolded by `rc init`. New projects are never created elsewhere.
	initRepoOwner = "rodolfochicone"
	// initTemplateRepo is the source template the new repository is created from.
	initTemplateRepo = "rodolfochicone/typescript-template"
)

// Sentinel errors describe recoverable GitHub configuration problems so callers
// (and tests) can match them with errors.Is. Each is wrapped together with an
// actionable, user-facing guidance block.
var (
	errGitHubCLIMissing       = errors.New("github cli (gh) not installed")
	errGitHubNotAuthenticated = errors.New("github cli (gh) not authenticated")
	errGitHubOrgAccess        = errors.New("github organization access denied")
)

const (
	guidanceGitHubCLIMissing = `Install the GitHub CLI and authenticate before running rc init:
  • Install: https://cli.github.com  (macOS: brew install gh)
  • Authenticate: gh auth login
  • Make sure your account can access the ` + initRepoOwner + ` organization.`

	guidanceGitHubNotAuthenticated = `Authenticate the GitHub CLI before running rc init:
  • Run: gh auth login
  • Then confirm access to the ` + initRepoOwner + ` organization:
      gh api user/memberships/orgs/` + initRepoOwner + ``

	guidanceGitHubOrgAccess = `Your account is not allowed to create repositories in the ` + initRepoOwner + ` organization:
  • Confirm membership: gh api user/memberships/orgs/` + initRepoOwner + `
  • Ask an organization admin to grant repository-creation permission.
  • Re-authenticate with org scope: gh auth login --scopes "repo,read:org"`
)

// ghRunner executes the gh CLI in dir (empty = current directory) and returns
// the combined stdout/stderr together with the process error. It is a field on
// initCommandState so tests can stub GitHub interactions.
type ghRunner func(ctx context.Context, dir string, args ...string) ([]byte, error)

type initCommandState struct {
	lookPath      func(string) (string, error)
	runGH         ghRunner
	getwd         func() (string, error)
	isInteractive func() bool
	promptName    func() (string, error)
}

func newInitCommandState() *initCommandState {
	return &initCommandState{
		lookPath:      exec.LookPath,
		runGH:         runGH,
		getwd:         os.Getwd,
		isInteractive: isInteractiveTerminal,
		promptName:    promptRepoName,
	}
}

// newInitCommand scaffolds a brand-new project: it creates a private repository
// in the rodolfochicone organization from the TypeScript template and clones it
// into a subdirectory of the current working directory.
func newInitCommand() *cobra.Command {
	state := newInitCommandState()
	cmd := &cobra.Command{
		Use:   "init <repo-name>",
		Short: "Scaffold a new project from the rodolfochicone TypeScript template",
		Long: `Create a new private repository in the ` + initRepoOwner + ` organization from the
` + initTemplateRepo + ` template, then clone it into ./<repo-name> under the current
directory.

Requires the GitHub CLI (gh) installed, authenticated, and with access to the
` + initRepoOwner + ` organization. When that is not the case, rc init explains how to
configure it.`,
		Example:      "  rc init my-new-service",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return state.run(cmd, args)
		},
	}
	return cmd
}

func (s *initCommandState) run(cmd *cobra.Command, args []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	repoName, err := s.resolveRepoName(args)
	if err != nil {
		return err
	}

	if err := s.ensureGitHubReady(ctx); err != nil {
		return err
	}

	workDir, err := s.getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	slug := initRepoOwner + "/" + repoName
	cmd.Printf("Creating %s from template %s...\n", slug, initTemplateRepo)

	output, err := s.runGH(ctx, workDir,
		"repo", "create", slug,
		"--template", initTemplateRepo,
		"--private",
		"--clone",
	)
	if err != nil {
		return explainCreateError(slug, output, err)
	}

	cmd.Printf("Created %s and cloned into ./%s\n", slug, repoName)
	return nil
}

// ensureGitHubReady verifies the GitHub CLI is installed and authenticated,
// returning actionable guidance when either precondition fails.
func (s *initCommandState) ensureGitHubReady(ctx context.Context) error {
	if _, err := s.lookPath("gh"); err != nil {
		return fmt.Errorf("%w\n\n%s", errGitHubCLIMissing, guidanceGitHubCLIMissing)
	}
	if _, err := s.runGH(ctx, "", "auth", "status"); err != nil {
		return fmt.Errorf("%w\n\n%s", errGitHubNotAuthenticated, guidanceGitHubNotAuthenticated)
	}
	return nil
}

// resolveRepoName takes the name from the positional argument when present,
// otherwise prompts for it interactively. In a non-interactive context with no
// argument it returns an actionable error instead of hanging on a prompt.
func (s *initCommandState) resolveRepoName(args []string) (string, error) {
	if len(args) == 1 {
		return validateRepoName(args[0])
	}
	if !s.isInteractive() {
		return "", errors.New("repository name is required: run `rc init <repo-name>`")
	}
	raw, err := s.promptName()
	if err != nil {
		return "", err
	}
	return validateRepoName(raw)
}

// promptRepoName asks the user for the new repository name, validating the
// input inline against the same rules as the positional argument.
func promptRepoName() (string, error) {
	var name string
	field := huh.NewInput().
		Key("repo").
		Title("New repository name").
		Placeholder("my-service").
		Description("Created as a private repository in the " + initRepoOwner + " organization.").
		Validate(func(value string) error {
			_, err := validateRepoName(value)
			return err
		}).
		Value(&name)
	if err := runPromptField(field); err != nil {
		return "", fmt.Errorf("prompt repository name: %w", err)
	}
	return name, nil
}

// validateRepoName rejects names that cannot be a plain repository slug, since
// the owner is always the rodolfochicone organization.
func validateRepoName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errors.New("repository name must not be empty")
	}
	if strings.ContainsAny(name, "/\\ \t") {
		return "", fmt.Errorf(
			"invalid repository name %q: use a plain name without slashes or spaces (it is always created under the %s org)",
			name,
			initRepoOwner,
		)
	}
	return name, nil
}

// explainCreateError turns a gh repo create failure into a clear message,
// attaching org-access guidance when the output looks like a permission denial.
func explainCreateError(slug string, output []byte, runErr error) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		detail = runErr.Error()
	}
	if isPermissionDenied(detail) {
		return fmt.Errorf("create %s: %s\n\n%w\n\n%s", slug, detail, errGitHubOrgAccess, guidanceGitHubOrgAccess)
	}
	return fmt.Errorf("create %s: %s", slug, detail)
}

func isPermissionDenied(detail string) bool {
	lower := strings.ToLower(detail)
	for _, marker := range []string{"permission", "not authorized", "must be a member", "403", "forbidden"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// runGH executes the gh CLI, capturing combined output for error reporting.
func runGH(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	return cmd.CombinedOutput()
}
