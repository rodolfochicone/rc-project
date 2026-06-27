package setup

import (
	"fmt"
	"io/fs"
)

// SyncConfig describes one bundled-skill sync run for a single agent.
//
// Sync reconciles the agent's installed bundled skills with the skills shipped
// in this binary: missing skills are added, drifted skills are updated, and
// skills that are already current are left untouched. Skills present in the
// agent directory that are not bundled with rc are never modified.
type SyncConfig struct {
	Bundle fs.FS

	ResolverOptions

	AgentName string
	Global    bool
	Mode      InstallMode
}

// SyncOutcome classifies what a sync run did to one bundled skill.
type SyncOutcome string

const (
	// SyncOutcomeAdded marks a skill that was missing and got installed.
	SyncOutcomeAdded SyncOutcome = "added"
	// SyncOutcomeUpdated marks an installed skill that drifted and was refreshed.
	SyncOutcomeUpdated SyncOutcome = "updated"
)

// SyncResult summarizes one bundled-skill sync run for a single agent.
type SyncResult struct {
	Agent     Agent
	Scope     InstallScope
	Mode      InstallMode
	Added     []SuccessItem
	Updated   []SuccessItem
	Unchanged []Skill
	Failed    []FailureItem
}

// Sync updates the bundled skills the agent already has and adds the ones it is
// missing, scoped to the selected project or global directory. Skills already
// current are skipped and non-bundled skills installed alongside them are left
// intact.
func Sync(cfg SyncConfig) (SyncResult, error) {
	if cfg.Bundle == nil {
		return SyncResult{}, fmt.Errorf("sync bundled skills: bundle is nil")
	}

	allSkills, err := ListSkills(cfg.Bundle)
	if err != nil {
		return SyncResult{}, err
	}

	allAgents, err := SupportedAgents(cfg.ResolverOptions)
	if err != nil {
		return SyncResult{}, err
	}
	selectedAgents, err := SelectAgents(allAgents, []string{cfg.AgentName})
	if err != nil {
		return SyncResult{}, err
	}
	agent := selectedAgents[0]

	env, err := resolveEnvironment(cfg.ResolverOptions)
	if err != nil {
		return SyncResult{}, err
	}

	entries, err := verificationEntries(allSkills, agent, env, cfg.Global)
	if err != nil {
		return SyncResult{}, err
	}

	scope := InstallScopeProject
	if cfg.Global {
		scope = InstallScopeGlobal
	}

	plan, err := classifySyncEntries(cfg.Bundle, scope, entries)
	if err != nil {
		return SyncResult{}, err
	}

	mode := cfg.Mode
	if mode == "" {
		mode = detectInstallMode(entries)
	}

	result := SyncResult{
		Agent:     agent,
		Scope:     scope,
		Mode:      mode,
		Unchanged: plan.unchanged,
	}
	if len(plan.toInstall) == 0 {
		return result, nil
	}

	successes, failures, err := InstallSelectedSkills(
		cfg.ResolverOptions,
		plan.toInstall,
		[]string{agent.Name},
		cfg.Global,
		mode,
	)
	if err != nil {
		return SyncResult{}, err
	}
	result.Failed = failures

	for i := range successes {
		switch plan.outcomes[successes[i].Skill.Name] {
		case SyncOutcomeAdded:
			result.Added = append(result.Added, successes[i])
		case SyncOutcomeUpdated:
			result.Updated = append(result.Updated, successes[i])
		}
	}
	return result, nil
}

// syncPlan captures the classification of bundled skills before any writes.
type syncPlan struct {
	toInstall []Skill
	unchanged []Skill
	outcomes  map[string]SyncOutcome
}

// classifySyncEntries sorts each bundled skill into add/update/skip buckets by
// comparing the installed copy in the selected scope against the bundle.
func classifySyncEntries(bundle fs.FS, scope InstallScope, entries []verificationEntry) (syncPlan, error) {
	plan := syncPlan{outcomes: make(map[string]SyncOutcome, len(entries))}
	for i := range entries {
		verified, err := verifyEntry(bundle, scope, entries[i])
		if err != nil {
			return syncPlan{}, err
		}
		switch verified.State {
		case VerifyStateMissing:
			plan.toInstall = append(plan.toInstall, entries[i].Skill)
			plan.outcomes[entries[i].Skill.Name] = SyncOutcomeAdded
		case VerifyStateDrifted:
			plan.toInstall = append(plan.toInstall, entries[i].Skill)
			plan.outcomes[entries[i].Skill.Name] = SyncOutcomeUpdated
		case VerifyStateCurrent:
			plan.unchanged = append(plan.unchanged, entries[i].Skill)
		}
	}
	return plan, nil
}
