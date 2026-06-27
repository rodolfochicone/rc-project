// Package configsvc implements ConfigService, a thin adapter over the workspace
// config subsystem that provides read/write access to global and per-workspace
// rc config files via the daemon HTTP API.
package configsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rodolfochicone/rc-project/internal/api/contract"
	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/core/workspace"
)

// Service adapts the workspace config subsystem for the daemon API handlers.
type Service struct {
	workspaces apicore.WorkspaceService
}

var _ apicore.ConfigService = (*Service)(nil)

// New constructs a Service. workspaces is used to resolve workspace root dirs for
// per-workspace config reads/writes.
func New(workspaces apicore.WorkspaceService) *Service {
	return &Service{workspaces: workspaces}
}

// GetGlobal reads the global (~/.rc/config.toml) rc config.
func (s *Service) GetGlobal(ctx context.Context) (contract.ConfigDocument, error) {
	cfg, _, err := workspace.LoadGlobalConfig(ctx)
	if err != nil {
		return contract.ConfigDocument{}, fmt.Errorf("get global config: %w", err)
	}
	return projectConfigToDocument(cfg), nil
}

// PutGlobal validates and atomically writes the global config, then re-reads it.
func (s *Service) PutGlobal(ctx context.Context, doc contract.ConfigDocument) (contract.ConfigDocument, error) {
	globalPath, err := workspace.GlobalConfigPath()
	if err != nil {
		return contract.ConfigDocument{}, fmt.Errorf("resolve global config path: %w", err)
	}

	cfg := documentToProjectConfig(doc)
	if err := workspace.WriteConfig(ctx, globalPath, cfg, workspace.GlobalConfigScope); err != nil {
		return contract.ConfigDocument{}, wrapValidationError(err)
	}

	result, _, err := workspace.LoadGlobalConfig(ctx)
	if err != nil {
		return contract.ConfigDocument{}, fmt.Errorf("reload global config after write: %w", err)
	}
	return projectConfigToDocument(result), nil
}

// GetWorkspace reads the per-workspace (.rc/config.toml) rc config.
func (s *Service) GetWorkspace(ctx context.Context, workspaceID string) (contract.ConfigDocument, error) {
	root, err := s.workspaceRoot(ctx, workspaceID)
	if err != nil {
		return contract.ConfigDocument{}, err
	}
	cfg, _, err := workspace.LoadConfig(ctx, root)
	if err != nil {
		return contract.ConfigDocument{}, fmt.Errorf("get workspace config: %w", err)
	}
	return projectConfigToDocument(cfg), nil
}

// PutWorkspace validates and atomically writes the per-workspace config, then re-reads it.
func (s *Service) PutWorkspace(
	ctx context.Context,
	workspaceID string,
	doc contract.ConfigDocument,
) (contract.ConfigDocument, error) {
	root, err := s.workspaceRoot(ctx, workspaceID)
	if err != nil {
		return contract.ConfigDocument{}, err
	}

	configPath := model.ConfigPathForWorkspace(root)
	cfg := documentToProjectConfig(doc)
	if err := workspace.WriteConfig(ctx, configPath, cfg, workspace.WorkspaceConfigScope); err != nil {
		return contract.ConfigDocument{}, wrapValidationError(err)
	}

	result, _, err := workspace.LoadConfig(ctx, root)
	if err != nil {
		return contract.ConfigDocument{}, fmt.Errorf("reload workspace config after write: %w", err)
	}
	return projectConfigToDocument(result), nil
}

func (s *Service) workspaceRoot(ctx context.Context, workspaceID string) (string, error) {
	ws, err := s.workspaces.Get(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("get workspace: %w", err)
	}
	return ws.RootDir, nil
}

// wrapValidationError converts workspace validation errors into a 400 config_invalid Problem.
// If err is already (or wraps) a *contract.Problem it is returned as-is.
func wrapValidationError(err error) error {
	if err == nil {
		return nil
	}
	var p *contract.Problem
	if errors.As(err, &p) {
		return err
	}
	return contract.NewProblem(http.StatusBadRequest, "config_invalid", err.Error(), nil, err)
}

// projectConfigToDocument converts the internal config type to the API document shape.
func projectConfigToDocument(cfg workspace.ProjectConfig) contract.ConfigDocument {
	return contract.ConfigDocument{
		Defaults:     defaultsToDoc(cfg.Defaults),
		Tasks:        tasksToDoc(cfg.Tasks),
		FixReviews:   fixReviewsToDoc(cfg.FixReviews),
		FetchReviews: fetchReviewsToDoc(cfg.FetchReviews),
		WatchReviews: watchReviewsToDoc(cfg.WatchReviews),
		Exec:         execToDoc(cfg.Exec),
		Runs:         runsToDoc(cfg.Runs),
		Sound:        soundToDoc(cfg.Sound),
	}
}

func defaultsToDoc(d workspace.DefaultsConfig) *contract.ConfigDefaults {
	if isZeroDefaults(d) {
		return nil
	}
	return &contract.ConfigDefaults{
		IDE: d.IDE, Model: d.Model, OutputFormat: d.OutputFormat,
		ReasoningEffort: d.ReasoningEffort, AccessMode: d.AccessMode,
		Timeout: d.Timeout, TailLines: d.TailLines, AddDirs: d.AddDirs,
		AutoCommit: d.AutoCommit, MaxRetries: d.MaxRetries,
		RetryBackoffMultiplier: d.RetryBackoffMultiplier,
	}
}

func tasksToDoc(t workspace.TasksConfig) *contract.ConfigTasks {
	if isZeroTasks(t) {
		return nil
	}
	ct := &contract.ConfigTasks{Types: t.Types}
	if r := t.Run; !isZeroTaskRun(r) {
		ct.Run = &contract.ConfigTaskRun{
			IncludeCompleted: r.IncludeCompleted,
			OutputFormat:     r.OutputFormat,
			TUI:              r.TUI,
			TaskRuntimeRules: taskRuleSliceToDoc(r.TaskRuntimeRules),
		}
	}
	return ct
}

func taskRuleSliceToDoc(rules *[]model.TaskRuntimeRule) *[]contract.ConfigTaskRuntimeRule {
	if rules == nil {
		return nil
	}
	out := make([]contract.ConfigTaskRuntimeRule, len(*rules))
	for i, r := range *rules {
		out[i] = contract.ConfigTaskRuntimeRule{
			ID:              r.ID,
			Type:            r.Type,
			IDE:             r.IDE,
			Model:           r.Model,
			ReasoningEffort: r.ReasoningEffort,
		}
	}
	return &out
}

func taskRuleSliceFromDoc(rules *[]contract.ConfigTaskRuntimeRule) *[]model.TaskRuntimeRule {
	if rules == nil {
		return nil
	}
	out := make([]model.TaskRuntimeRule, len(*rules))
	for i, r := range *rules {
		out[i] = model.TaskRuntimeRule{
			ID:              r.ID,
			Type:            r.Type,
			IDE:             r.IDE,
			Model:           r.Model,
			ReasoningEffort: r.ReasoningEffort,
		}
	}
	return &out
}

func fixReviewsToDoc(f workspace.FixReviewsConfig) *contract.ConfigFixReviews {
	if isZeroFixReviews(f) {
		return nil
	}
	return &contract.ConfigFixReviews{
		Concurrent: f.Concurrent, BatchSize: f.BatchSize,
		IncludeResolved: f.IncludeResolved, OutputFormat: f.OutputFormat, TUI: f.TUI,
	}
}

func fetchReviewsToDoc(f workspace.FetchReviewsConfig) *contract.ConfigFetchReviews {
	if isZeroFetchReviews(f) {
		return nil
	}
	return &contract.ConfigFetchReviews{Provider: f.Provider, Nitpicks: f.Nitpicks}
}

func watchReviewsToDoc(w workspace.WatchReviewsConfig) *contract.ConfigWatchReviews {
	if isZeroWatchReviews(w) {
		return nil
	}
	return &contract.ConfigWatchReviews{
		MaxRounds: w.MaxRounds, PollInterval: w.PollInterval,
		ReviewTimeout: w.ReviewTimeout, QuietPeriod: w.QuietPeriod,
		AutoPush: w.AutoPush, UntilClean: w.UntilClean,
		PushRemote: w.PushRemote, PushBranch: w.PushBranch,
	}
}

func execToDoc(e workspace.ExecConfig) *contract.ConfigExec {
	if isZeroExec(e) {
		return nil
	}
	return &contract.ConfigExec{
		IDE: e.IDE, Model: e.Model, OutputFormat: e.OutputFormat,
		ReasoningEffort: e.ReasoningEffort, AccessMode: e.AccessMode,
		Timeout: e.Timeout, TailLines: e.TailLines, AddDirs: e.AddDirs,
		AutoCommit: e.AutoCommit, MaxRetries: e.MaxRetries,
		RetryBackoffMultiplier: e.RetryBackoffMultiplier,
		Verbose:                e.Verbose, TUI: e.TUI, Persist: e.Persist,
	}
}

func runsToDoc(r workspace.RunsConfig) *contract.ConfigRuns {
	if isZeroRuns(r) {
		return nil
	}
	return &contract.ConfigRuns{
		DefaultAttachMode: r.DefaultAttachMode, KeepTerminalDays: r.KeepTerminalDays,
		KeepMax: r.KeepMax, ShutdownDrainTimeout: r.ShutdownDrainTimeout,
	}
}

func soundToDoc(s workspace.SoundConfig) *contract.ConfigSound {
	if isZeroSound(s) {
		return nil
	}
	return &contract.ConfigSound{Enabled: s.Enabled, OnCompleted: s.OnCompleted, OnFailed: s.OnFailed}
}

// documentToProjectConfig converts the API document shape to the internal config type.
func documentToProjectConfig(doc contract.ConfigDocument) workspace.ProjectConfig {
	return workspace.ProjectConfig{
		Defaults:     docToDefaults(doc.Defaults),
		Tasks:        docToTasks(doc.Tasks),
		FixReviews:   docToFixReviews(doc.FixReviews),
		FetchReviews: docToFetchReviews(doc.FetchReviews),
		WatchReviews: docToWatchReviews(doc.WatchReviews),
		Exec:         docToExec(doc.Exec),
		Runs:         docToRuns(doc.Runs),
		Sound:        docToSound(doc.Sound),
	}
}

func docToDefaults(d *contract.ConfigDefaults) workspace.DefaultsConfig {
	if d == nil {
		return workspace.DefaultsConfig{}
	}
	return workspace.DefaultsConfig{
		IDE: d.IDE, Model: d.Model, OutputFormat: d.OutputFormat,
		ReasoningEffort: d.ReasoningEffort, AccessMode: d.AccessMode,
		Timeout: d.Timeout, TailLines: d.TailLines, AddDirs: d.AddDirs,
		AutoCommit: d.AutoCommit, MaxRetries: d.MaxRetries,
		RetryBackoffMultiplier: d.RetryBackoffMultiplier,
	}
}

func docToTasks(t *contract.ConfigTasks) workspace.TasksConfig {
	if t == nil {
		return workspace.TasksConfig{}
	}
	tc := workspace.TasksConfig{Types: t.Types}
	if r := t.Run; r != nil {
		tc.Run = workspace.TaskRunConfig{
			IncludeCompleted: r.IncludeCompleted,
			OutputFormat:     r.OutputFormat,
			TUI:              r.TUI,
			TaskRuntimeRules: taskRuleSliceFromDoc(r.TaskRuntimeRules),
		}
	}
	return tc
}

func docToFixReviews(f *contract.ConfigFixReviews) workspace.FixReviewsConfig {
	if f == nil {
		return workspace.FixReviewsConfig{}
	}
	return workspace.FixReviewsConfig{
		Concurrent: f.Concurrent, BatchSize: f.BatchSize,
		IncludeResolved: f.IncludeResolved, OutputFormat: f.OutputFormat, TUI: f.TUI,
	}
}

func docToFetchReviews(f *contract.ConfigFetchReviews) workspace.FetchReviewsConfig {
	if f == nil {
		return workspace.FetchReviewsConfig{}
	}
	return workspace.FetchReviewsConfig{Provider: f.Provider, Nitpicks: f.Nitpicks}
}

func docToWatchReviews(w *contract.ConfigWatchReviews) workspace.WatchReviewsConfig {
	if w == nil {
		return workspace.WatchReviewsConfig{}
	}
	return workspace.WatchReviewsConfig{
		MaxRounds: w.MaxRounds, PollInterval: w.PollInterval,
		ReviewTimeout: w.ReviewTimeout, QuietPeriod: w.QuietPeriod,
		AutoPush: w.AutoPush, UntilClean: w.UntilClean,
		PushRemote: w.PushRemote, PushBranch: w.PushBranch,
	}
}

func docToExec(e *contract.ConfigExec) workspace.ExecConfig {
	if e == nil {
		return workspace.ExecConfig{}
	}
	return workspace.ExecConfig{
		RuntimeOverrides: workspace.RuntimeOverrides{
			IDE: e.IDE, Model: e.Model, OutputFormat: e.OutputFormat,
			ReasoningEffort: e.ReasoningEffort, AccessMode: e.AccessMode,
			Timeout: e.Timeout, TailLines: e.TailLines, AddDirs: e.AddDirs,
			AutoCommit: e.AutoCommit, MaxRetries: e.MaxRetries,
			RetryBackoffMultiplier: e.RetryBackoffMultiplier,
		},
		Verbose: e.Verbose, TUI: e.TUI, Persist: e.Persist,
	}
}

func docToRuns(r *contract.ConfigRuns) workspace.RunsConfig {
	if r == nil {
		return workspace.RunsConfig{}
	}
	return workspace.RunsConfig{
		DefaultAttachMode: r.DefaultAttachMode, KeepTerminalDays: r.KeepTerminalDays,
		KeepMax: r.KeepMax, ShutdownDrainTimeout: r.ShutdownDrainTimeout,
	}
}

func docToSound(s *contract.ConfigSound) workspace.SoundConfig {
	if s == nil {
		return workspace.SoundConfig{}
	}
	return workspace.SoundConfig{Enabled: s.Enabled, OnCompleted: s.OnCompleted, OnFailed: s.OnFailed}
}

// zero-check helpers — each returns true when all pointer fields are nil,
// meaning the sub-config section is absent from the document.

func isZeroDefaults(d workspace.DefaultsConfig) bool {
	return d.IDE == nil && d.Model == nil && d.OutputFormat == nil &&
		d.ReasoningEffort == nil && d.AccessMode == nil && d.Timeout == nil &&
		d.TailLines == nil && d.AddDirs == nil && d.AutoCommit == nil &&
		d.MaxRetries == nil && d.RetryBackoffMultiplier == nil
}

func isZeroTaskRun(r workspace.TaskRunConfig) bool {
	return r.IncludeCompleted == nil && r.OutputFormat == nil && r.TUI == nil && r.TaskRuntimeRules == nil
}

func isZeroTasks(t workspace.TasksConfig) bool {
	return t.Types == nil && isZeroTaskRun(t.Run)
}

func isZeroFixReviews(f workspace.FixReviewsConfig) bool {
	return f.Concurrent == nil && f.BatchSize == nil && f.IncludeResolved == nil &&
		f.OutputFormat == nil && f.TUI == nil
}

func isZeroFetchReviews(f workspace.FetchReviewsConfig) bool {
	return f.Provider == nil && f.Nitpicks == nil
}

func isZeroWatchReviews(w workspace.WatchReviewsConfig) bool {
	return w.MaxRounds == nil && w.PollInterval == nil && w.ReviewTimeout == nil &&
		w.QuietPeriod == nil && w.AutoPush == nil && w.UntilClean == nil &&
		w.PushRemote == nil && w.PushBranch == nil
}

func isZeroExec(e workspace.ExecConfig) bool {
	return isZeroDefaults(workspace.DefaultsConfig(e.RuntimeOverrides)) &&
		e.Verbose == nil && e.TUI == nil && e.Persist == nil
}

func isZeroRuns(r workspace.RunsConfig) bool {
	return r.DefaultAttachMode == nil && r.KeepTerminalDays == nil &&
		r.KeepMax == nil && r.ShutdownDrainTimeout == nil
}

func isZeroSound(s workspace.SoundConfig) bool {
	return s.Enabled == nil && s.OnCompleted == nil && s.OnFailed == nil
}
