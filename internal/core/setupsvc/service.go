// Package setupsvc implements apicore.SetupService, a thin adapter that lets
// the daemon HTTP API detect installable agents and install bundled rc skills
// into a project directory (project-scoped, never global).
package setupsvc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rodolfochicone/rc-project/internal/api/contract"
	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/setup"
)

// Service implements apicore.SetupService by delegating to the internal/setup
// package. The setup functions are held as fields so tests can inject stubs.
type Service struct {
	supportedAgents func(setup.ResolverOptions) ([]setup.Agent, error)
	detectAgents    func(setup.ResolverOptions) ([]setup.Agent, error)
	listSkills      func() ([]setup.Skill, error)
	preview         func(setup.InstallConfig) ([]setup.PreviewItem, error)
	install         func(setup.InstallConfig) (*setup.Result, error)
	resolver        func() setup.ResolverOptions
}

var _ apicore.SetupService = (*Service)(nil)

// Option customizes a Service, primarily for test injection.
type Option func(*Service)

// New constructs a Service backed by the production setup package.
func New(opts ...Option) *Service {
	s := &Service{
		supportedAgents: setup.SupportedAgents,
		detectAgents:    setup.DetectInstalledAgents,
		listSkills:      setup.ListBundledSkills,
		preview:         setup.PreviewBundledSkillInstall,
		install:         setup.InstallBundledSetupAssets,
		resolver:        defaultResolverOptions,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// defaultResolverOptions mirrors the CLI's environment-sensitive resolver so the
// daemon installs to the same agent paths a terminal `rc setup` would.
func defaultResolverOptions() setup.ResolverOptions {
	return setup.ResolverOptions{
		CodeXHome:       strings.TrimSpace(os.Getenv("CODEX_HOME")),
		ClaudeConfigDir: strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR")),
		XDGConfigHome:   strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")),
	}
}

// Options lists every supported agent (flagged with whether it is detected on
// this machine), the bundled rc skills available for installation, and whether
// the project is already configured (every bundled skill present for at least
// one detected agent, in project or global scope).
func (s *Service) Options(
	_ context.Context,
	projectRoot string,
) (contract.SetupOptionsResponse, error) {
	resolver := s.resolver()
	resolver.CWD = strings.TrimSpace(projectRoot)

	supported, err := s.supportedAgents(resolver)
	if err != nil {
		return contract.SetupOptionsResponse{}, fmt.Errorf("list setup agents: %w", err)
	}
	detected, err := s.detectAgents(resolver)
	if err != nil {
		return contract.SetupOptionsResponse{}, fmt.Errorf("detect setup agents: %w", err)
	}
	detectedNames := make(map[string]struct{}, len(detected))
	for i := range detected {
		detectedNames[detected[i].Name] = struct{}{}
	}

	agents := make([]contract.SetupAgent, 0, len(supported))
	for i := range supported {
		_, isDetected := detectedNames[supported[i].Name]
		agents = append(agents, contract.SetupAgent{
			Name:        supported[i].Name,
			DisplayName: supported[i].DisplayName,
			Detected:    isDetected,
		})
	}

	skills, err := s.listSkills()
	if err != nil {
		return contract.SetupOptionsResponse{}, fmt.Errorf("list bundled skills: %w", err)
	}
	items := make([]contract.SetupSkill, 0, len(skills))
	skillNames := make([]string, 0, len(skills))
	for i := range skills {
		items = append(items, contract.SetupSkill{
			Name:        skills[i].Name,
			Description: skills[i].Description,
		})
		skillNames = append(skillNames, skills[i].Name)
	}

	configured, err := s.isConfigured(resolver, detected, skillNames)
	if err != nil {
		return contract.SetupOptionsResponse{}, err
	}

	return contract.SetupOptionsResponse{Agents: agents, Skills: items, Configured: configured}, nil
}

// isConfigured reports whether at least one detected agent already has every
// bundled skill present, checking project scope first and then global scope so
// users who installed skills globally are not prompted per project.
func (s *Service) isConfigured(
	resolver setup.ResolverOptions,
	detected []setup.Agent,
	skillNames []string,
) (bool, error) {
	if len(detected) == 0 || len(skillNames) == 0 {
		return false, nil
	}

	agentNames := make([]string, 0, len(detected))
	for i := range detected {
		agentNames = append(agentNames, detected[i].Name)
	}

	// present[agent][skill] is true when the skill is installed in either scope.
	present := make(map[string]map[string]bool, len(agentNames))
	for _, global := range []bool{false, true} {
		items, err := s.preview(setup.InstallConfig{
			ResolverOptions: resolver,
			SkillNames:      skillNames,
			AgentNames:      agentNames,
			Global:          global,
		})
		if err != nil {
			return false, fmt.Errorf("preview setup install: %w", err)
		}
		for i := range items {
			if !items[i].WillOverwrite {
				continue
			}
			agent := items[i].Agent.Name
			if present[agent] == nil {
				present[agent] = make(map[string]bool, len(skillNames))
			}
			present[agent][items[i].Skill.Name] = true
		}
	}

	for _, agent := range agentNames {
		if len(present[agent]) == len(skillNames) {
			return true, nil
		}
	}
	return false, nil
}

// Install copies the selected bundled skills into the project directory for the
// selected agents. An empty skill selection installs every bundled skill.
func (s *Service) Install(
	_ context.Context,
	projectRoot string,
	agents, skills []string,
) (contract.SetupInstallResponse, error) {
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		return contract.SetupInstallResponse{}, errors.New("setup install: project root is required")
	}
	if len(agents) == 0 {
		return contract.SetupInstallResponse{}, errors.New("setup install: at least one agent is required")
	}

	if len(skills) == 0 {
		all, err := s.listSkills()
		if err != nil {
			return contract.SetupInstallResponse{}, fmt.Errorf("list bundled skills: %w", err)
		}
		skills = make([]string, 0, len(all))
		for i := range all {
			skills = append(skills, all[i].Name)
		}
	}

	resolver := s.resolver()
	resolver.CWD = root

	result, err := s.install(setup.InstallConfig{
		ResolverOptions: resolver,
		SkillNames:      skills,
		AgentNames:      agents,
		Global:          false,
	})
	if err != nil {
		return contract.SetupInstallResponse{}, fmt.Errorf("install bundled skills: %w", err)
	}

	response := contract.SetupInstallResponse{}
	for i := range result.Successful {
		item := &result.Successful[i]
		response.Installed = append(response.Installed, contract.SetupInstalledItem{
			Skill: item.Skill.Name,
			Agent: item.Agent.DisplayName,
			Path:  item.Path,
		})
	}
	for i := range result.Failed {
		item := &result.Failed[i]
		response.Failed = append(response.Failed, contract.SetupFailedItem{
			Skill: item.Skill.Name,
			Agent: item.Agent.DisplayName,
			Error: item.Error,
		})
	}
	return response, nil
}
