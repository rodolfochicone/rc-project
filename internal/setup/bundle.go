package setup

import (
	"fmt"
	"io/fs"

	"github.com/rodolfochicone/rc-project/agents"
	"github.com/rodolfochicone/rc-project/skills"
)

var installBundledReusableAgents = InstallBundledReusableAgents
var installBundledCommands = InstallBundledCommands
var installBundledHooks = InstallBundledHooks

// ListBundledSkills returns the public skills bundled into the rc binary.
func ListBundledSkills() ([]Skill, error) {
	return ListSkills(skills.FS)
}

// PreviewBundledSkillInstall resolves the on-disk install plan for bundled skills.
func PreviewBundledSkillInstall(cfg InstallConfig) ([]PreviewItem, error) {
	cfg.Bundle = skills.FS
	return Preview(cfg)
}

// InstallBundledSkills materializes bundled public skills for the selected agents.
func InstallBundledSkills(cfg InstallConfig) (*Result, error) {
	cfg.Bundle = skills.FS
	return Install(cfg)
}

// InstallBundledSetupAssets materializes bundled skills and any bundled reusable agents.
func InstallBundledSetupAssets(cfg InstallConfig) (*Result, error) {
	cfg.Bundle = skills.FS
	result, err := Install(cfg)
	if err != nil {
		return nil, err
	}

	successes, failures, err := installBundledReusableAgents(ReusableAgentInstallConfig{
		ResolverOptions: cfg.ResolverOptions,
		Global:          cfg.Global,
	})
	if err != nil {
		return result, fmt.Errorf("install bundled reusable agents: %w", err)
	}
	result.ReusableAgentsSuccessful = append(result.ReusableAgentsSuccessful, successes...)
	result.ReusableAgentsFailed = append(result.ReusableAgentsFailed, failures...)

	cmdSuccesses, cmdFailures, err := installBundledCommands(CommandInstallConfig{
		ResolverOptions: cfg.ResolverOptions,
		Global:          cfg.Global,
	})
	if err != nil {
		return result, fmt.Errorf("install bundled commands: %w", err)
	}
	result.CommandsSuccessful = append(result.CommandsSuccessful, cmdSuccesses...)
	result.CommandsFailed = append(result.CommandsFailed, cmdFailures...)

	hookSuccesses, hookFailures, err := installBundledHooks(HookInstallConfig{
		ResolverOptions: cfg.ResolverOptions,
		Global:          cfg.Global,
	})
	if err != nil {
		return result, fmt.Errorf("install bundled hooks: %w", err)
	}
	result.HooksSuccessful = append(result.HooksSuccessful, hookSuccesses...)
	result.HooksFailed = append(result.HooksFailed, hookFailures...)
	return result, nil
}

// VerifyBundledSkills checks whether bundled public skills are installed and current.
func VerifyBundledSkills(cfg VerifyConfig) (VerifyResult, error) {
	cfg.Bundle = skills.FS
	return Verify(cfg)
}

// SyncBundledSkills updates the bundled public skills the agent already has and
// adds the ones it is missing for the selected scope.
func SyncBundledSkills(cfg SyncConfig) (SyncResult, error) {
	cfg.Bundle = skills.FS
	return Sync(cfg)
}

// bundledSkillsRoot returns the embedded skill filesystem for tests.
func bundledSkillsRoot() (fs.FS, error) {
	return skills.FS, nil
}

// bundledReusableAgentsRoot returns the embedded reusable-agent filesystem for tests.
func bundledReusableAgentsRoot() (fs.FS, error) {
	return agents.FS, nil
}
