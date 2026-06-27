package setupsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/setup"
)

// withStubs returns options that replace every external setup dependency so the
// tests exercise the service's own mapping, validation, and scope logic.
func withStubs(
	supported, detected []setup.Agent,
	skills []setup.Skill,
	install func(setup.InstallConfig) (*setup.Result, error),
) []Option {
	return []Option{
		func(s *Service) {
			s.supportedAgents = func(setup.ResolverOptions) ([]setup.Agent, error) { return supported, nil }
		},
		func(s *Service) {
			s.detectAgents = func(setup.ResolverOptions) ([]setup.Agent, error) { return detected, nil }
		},
		func(s *Service) { s.listSkills = func() ([]setup.Skill, error) { return skills, nil } },
		// Default: nothing installed yet (project is unconfigured).
		func(s *Service) {
			s.preview = func(setup.InstallConfig) ([]setup.PreviewItem, error) { return nil, nil }
		},
		func(s *Service) { s.install = install },
		func(s *Service) { s.resolver = func() setup.ResolverOptions { return setup.ResolverOptions{} } },
	}
}

// withPreview overrides the preview seam after the defaults from withStubs.
func withPreview(preview func(setup.InstallConfig) ([]setup.PreviewItem, error)) Option {
	return func(s *Service) { s.preview = preview }
}

func TestOptionsFlagsDetectedAgentsAndMapsSkills(t *testing.T) {
	t.Parallel()

	svc := New(withStubs(
		[]setup.Agent{
			{Name: "claude", DisplayName: "Claude"},
			{Name: "codex", DisplayName: "Codex"},
		},
		[]setup.Agent{{Name: "claude", DisplayName: "Claude"}},
		[]setup.Skill{{Name: "rc-create-prd", Description: "Write a PRD"}},
		func(setup.InstallConfig) (*setup.Result, error) { return &setup.Result{}, nil },
	)...)

	got, err := svc.Options(context.Background(), "/proj")
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}

	// Nothing installed (default preview stub) → project is not configured.
	if got.Configured {
		t.Error("Configured = true, want false for an empty project")
	}

	if len(got.Agents) != 2 {
		t.Fatalf("len(agents) = %d, want 2", len(got.Agents))
	}
	// Detection drives the UI default selection, so the flag must be accurate.
	byName := map[string]bool{}
	for _, agent := range got.Agents {
		byName[agent.Name] = agent.Detected
	}
	if !byName["claude"] {
		t.Error("claude should be flagged detected")
	}
	if byName["codex"] {
		t.Error("codex should not be flagged detected")
	}

	if len(got.Skills) != 1 || got.Skills[0].Name != "rc-create-prd" ||
		got.Skills[0].Description != "Write a PRD" {
		t.Fatalf("skills = %+v, want one mapped rc-create-prd", got.Skills)
	}
}

func TestOptionsReportsConfiguredWhenDetectedAgentHasAllSkills(t *testing.T) {
	t.Parallel()

	skills := []setup.Skill{{Name: "rc-create-prd"}, {Name: "rc-final-verify"}}
	opts := append(
		withStubs(
			[]setup.Agent{{Name: "claude-code", DisplayName: "Claude Code"}},
			[]setup.Agent{{Name: "claude-code", DisplayName: "Claude Code"}},
			skills,
			func(setup.InstallConfig) (*setup.Result, error) { return &setup.Result{}, nil },
		),
		// Project scope reports both skills already present for the detected agent.
		withPreview(func(cfg setup.InstallConfig) ([]setup.PreviewItem, error) {
			if cfg.Global {
				return nil, nil
			}
			return []setup.PreviewItem{
				{
					Skill:         setup.Skill{Name: "rc-create-prd"},
					Agent:         setup.Agent{Name: "claude-code"},
					WillOverwrite: true,
				},
				{
					Skill:         setup.Skill{Name: "rc-final-verify"},
					Agent:         setup.Agent{Name: "claude-code"},
					WillOverwrite: true,
				},
			}, nil
		}),
	)

	got, err := New(opts...).Options(context.Background(), "/proj")
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if !got.Configured {
		t.Error("Configured = false, want true when all skills are installed for a detected agent")
	}
}

func TestOptionsNotConfiguredWhenSkillsOnlyPartiallyInstalled(t *testing.T) {
	t.Parallel()

	skills := []setup.Skill{{Name: "rc-create-prd"}, {Name: "rc-final-verify"}}
	opts := append(
		withStubs(
			[]setup.Agent{{Name: "claude-code", DisplayName: "Claude Code"}},
			[]setup.Agent{{Name: "claude-code", DisplayName: "Claude Code"}},
			skills,
			func(setup.InstallConfig) (*setup.Result, error) { return &setup.Result{}, nil },
		),
		// Only one of two skills present → still needs configuration.
		withPreview(func(cfg setup.InstallConfig) ([]setup.PreviewItem, error) {
			if cfg.Global {
				return nil, nil
			}
			return []setup.PreviewItem{
				{
					Skill:         setup.Skill{Name: "rc-create-prd"},
					Agent:         setup.Agent{Name: "claude-code"},
					WillOverwrite: true,
				},
			}, nil
		}),
	)

	got, err := New(opts...).Options(context.Background(), "/proj")
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if got.Configured {
		t.Error("Configured = true, want false when only some skills are installed")
	}
}

func TestInstallRejectsMissingInputs(t *testing.T) {
	t.Parallel()

	svc := New(withStubs(nil, nil, nil,
		func(setup.InstallConfig) (*setup.Result, error) {
			t.Fatal("install must not run when validation fails")
			return nil, nil
		},
	)...)

	tests := []struct {
		name   string
		root   string
		agents []string
	}{
		{name: "blank project root", root: "   ", agents: []string{"claude"}},
		{name: "no agents", root: "/proj", agents: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := svc.Install(context.Background(), tc.root, tc.agents, []string{"rc-create-prd"}); err == nil {
				t.Fatal("Install() error = nil, want validation error")
			}
		})
	}
}

func TestInstallScopesToProjectAndDefaultsAllSkills(t *testing.T) {
	t.Parallel()

	var captured setup.InstallConfig
	svc := New(withStubs(
		nil, nil,
		[]setup.Skill{{Name: "rc-create-prd"}, {Name: "rc-final-verify"}},
		func(cfg setup.InstallConfig) (*setup.Result, error) {
			captured = cfg
			return &setup.Result{
				Successful: []setup.SuccessItem{
					{
						Skill: setup.Skill{Name: "rc-create-prd"},
						Agent: setup.Agent{DisplayName: "Claude"},
						Path:  "/proj/.claude/skills/rc-create-prd",
					},
				},
				Failed: []setup.FailureItem{
					{
						Skill: setup.Skill{Name: "rc-final-verify"},
						Agent: setup.Agent{DisplayName: "Claude"},
						Error: "permission denied",
					},
				},
			}, nil
		},
	)...)

	got, err := svc.Install(context.Background(), "/proj", []string{"claude"}, nil)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Project scope is the whole point: the install must target the workspace
	// directory and never the global scope.
	if captured.CWD != "/proj" {
		t.Errorf("install CWD = %q, want /proj", captured.CWD)
	}
	if captured.Global {
		t.Error("install must be project-scoped, got Global=true")
	}
	if captured.AgentNames[0] != "claude" {
		t.Errorf("install AgentNames = %v, want [claude]", captured.AgentNames)
	}
	// An empty selection installs every bundled skill.
	if len(captured.SkillNames) != 2 {
		t.Errorf("install SkillNames = %v, want all bundled skills", captured.SkillNames)
	}

	if len(got.Installed) != 1 || got.Installed[0].Skill != "rc-create-prd" {
		t.Errorf("Installed = %+v, want one rc-create-prd entry", got.Installed)
	}
	if len(got.Failed) != 1 || got.Failed[0].Error != "permission denied" {
		t.Errorf("Failed = %+v, want one permission-denied entry", got.Failed)
	}
}

func TestInstallPropagatesInstallerError(t *testing.T) {
	t.Parallel()

	svc := New(withStubs(nil, nil,
		[]setup.Skill{{Name: "rc-create-prd"}},
		func(setup.InstallConfig) (*setup.Result, error) {
			return nil, errors.New("disk full")
		},
	)...)

	if _, err := svc.Install(context.Background(), "/proj", []string{"claude"}, []string{"rc-create-prd"}); err == nil {
		t.Fatal("Install() error = nil, want propagated installer error")
	}
}
