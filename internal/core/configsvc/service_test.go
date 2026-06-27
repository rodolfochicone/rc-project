package configsvc_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/rodolfochicone/rc-project/internal/api/contract"
	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/core/configsvc"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/core/workspace"
)

// stubWorkspaceService is a minimal WorkspaceService that returns a fixed workspace.
type stubWorkspaceService struct {
	rootDir string
	err     error
}

var _ apicore.WorkspaceService = (*stubWorkspaceService)(nil)

func (s *stubWorkspaceService) Get(_ context.Context, _ string) (apicore.Workspace, error) {
	if s.err != nil {
		return apicore.Workspace{}, s.err
	}
	return apicore.Workspace{ID: "ws-1", RootDir: s.rootDir}, nil
}

func (s *stubWorkspaceService) Register(_ context.Context, _, _ string) (apicore.WorkspaceRegisterResult, error) {
	return apicore.WorkspaceRegisterResult{}, nil
}
func (s *stubWorkspaceService) List(_ context.Context) ([]apicore.Workspace, error) {
	return nil, nil
}

func (s *stubWorkspaceService) Update(
	_ context.Context,
	_ string,
	_ apicore.WorkspaceUpdateInput,
) (apicore.Workspace, error) {
	return apicore.Workspace{}, nil
}
func (s *stubWorkspaceService) Delete(_ context.Context, _ string) error { return nil }
func (s *stubWorkspaceService) Resolve(_ context.Context, _ string) (apicore.Workspace, error) {
	return apicore.Workspace{}, nil
}
func (s *stubWorkspaceService) Sync(_ context.Context) (apicore.WorkspaceSyncResult, error) {
	return apicore.WorkspaceSyncResult{}, nil
}

// TestGetWorkspaceConfigMapsProjectConfig asserts that GetWorkspace reads the
// on-disk config and returns a ConfigDocument whose optional fields match the
// written config. This matters because a silent mapping error would cause the
// UI to show wrong defaults to the user.
func TestGetWorkspaceConfigMapsProjectConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rcDir := filepath.Join(dir, ".rc")
	if err := os.MkdirAll(rcDir, 0o700); err != nil {
		t.Fatalf("mkdir .rc: %v", err)
	}

	keepMax := 42
	cfg := workspace.ProjectConfig{
		Runs: workspace.RunsConfig{KeepMax: &keepMax},
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		t.Fatalf("toml.Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rcDir, "config.toml"), data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	svc := configsvc.New(&stubWorkspaceService{rootDir: dir})
	doc, err := svc.GetWorkspace(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("GetWorkspace() error = %v", err)
	}
	if doc.Runs == nil {
		t.Fatal("GetWorkspace() Runs = nil, want non-nil")
	}
	if doc.Runs.KeepMax == nil || *doc.Runs.KeepMax != keepMax {
		t.Fatalf("GetWorkspace() Runs.KeepMax = %v, want %d", doc.Runs.KeepMax, keepMax)
	}
}

// TestPutWorkspaceConfigRoundTrip asserts that writing a ConfigDocument and then
// reading it back yields the same document. This matters because the daemon must
// persist exactly what the user submitted, not a silently mutated copy.
func TestPutWorkspaceConfigRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rcDir := filepath.Join(dir, ".rc")
	if err := os.MkdirAll(rcDir, 0o700); err != nil {
		t.Fatalf("mkdir .rc: %v", err)
	}

	keepMax := 7
	doc := contract.ConfigDocument{
		Runs: &contract.ConfigRuns{KeepMax: &keepMax},
	}

	svc := configsvc.New(&stubWorkspaceService{rootDir: dir})
	written, err := svc.PutWorkspace(context.Background(), "ws-1", doc)
	if err != nil {
		t.Fatalf("PutWorkspace() error = %v", err)
	}
	if written.Runs == nil || written.Runs.KeepMax == nil || *written.Runs.KeepMax != keepMax {
		t.Fatalf("PutWorkspace() written Runs.KeepMax = %v, want %d", written.Runs, keepMax)
	}

	reread, err := svc.GetWorkspace(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("GetWorkspace() after write error = %v", err)
	}
	if reread.Runs == nil || reread.Runs.KeepMax == nil || *reread.Runs.KeepMax != keepMax {
		t.Fatalf("GetWorkspace() after write Runs.KeepMax = %v, want %d", reread.Runs, keepMax)
	}
}

// TestPutWorkspaceConfigValidationReturnsConfigInvalid asserts that submitting
// an invalid config document returns a 400 config_invalid Problem and leaves the
// on-disk file unchanged. This matters because a daemon that persists invalid
// config would break on the next restart.
func TestPutWorkspaceConfigValidationReturnsConfigInvalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rcDir := filepath.Join(dir, ".rc")
	if err := os.MkdirAll(rcDir, 0o700); err != nil {
		t.Fatalf("mkdir .rc: %v", err)
	}

	// Write a valid config first so we can verify it is untouched after failure.
	keepMax := 3
	original := workspace.ProjectConfig{Runs: workspace.RunsConfig{KeepMax: &keepMax}}
	origData, _ := toml.Marshal(original)
	configPath := filepath.Join(rcDir, "config.toml")
	if err := os.WriteFile(configPath, origData, 0o600); err != nil {
		t.Fatalf("WriteFile seed: %v", err)
	}

	emptyProvider := ""
	invalid := contract.ConfigDocument{
		FetchReviews: &contract.ConfigFetchReviews{Provider: &emptyProvider},
	}

	svc := configsvc.New(&stubWorkspaceService{rootDir: dir})
	_, err := svc.PutWorkspace(context.Background(), "ws-1", invalid)
	if err == nil {
		t.Fatal("PutWorkspace() invalid config: want error, got nil")
	}

	var problem *contract.Problem
	if !errors.As(err, &problem) {
		t.Fatalf("PutWorkspace() error type = %T, want *contract.Problem", err)
	}
	if problem.Code != "config_invalid" {
		t.Fatalf("PutWorkspace() error code = %q, want config_invalid", problem.Code)
	}
	if problem.Status != http.StatusBadRequest {
		t.Fatalf("PutWorkspace() error status = %d, want 400", problem.Status)
	}

	// The on-disk file must be unchanged.
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile after failed put: %v", err)
	}
	if !bytes.Equal(got, origData) {
		t.Fatalf("config file was modified despite validation failure")
	}
}

// TestGetWorkspaceConfigWorkspaceServiceError asserts that a workspace service
// failure is propagated as an error (not swallowed). This matters because a
// silent failure would return an empty config to the user instead of an error.
func TestGetWorkspaceConfigWorkspaceServiceError(t *testing.T) {
	t.Parallel()

	boom := errors.New("workspace store offline")
	svc := configsvc.New(&stubWorkspaceService{err: boom})
	_, err := svc.GetWorkspace(context.Background(), "ws-1")
	if err == nil {
		t.Fatal("GetWorkspace() with workspace error: want error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("GetWorkspace() error = %v, want to wrap %v", err, boom)
	}
}

// TestPutGlobalConfigRoundTrip asserts that PutGlobal writes the config to the
// same path that GetGlobal reads from. This matters because a path divergence
// (e.g. PutGlobal writing to a different home dir than GetGlobal reads from)
// would silently discard every global config save the user makes in the UI.
func TestPutGlobalConfigRoundTrip(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	if err := os.MkdirAll(filepath.Join(fakeHome, ".rc"), 0o700); err != nil {
		t.Fatalf("mkdir .rc: %v", err)
	}

	keepMax := 55
	doc := contract.ConfigDocument{
		Runs: &contract.ConfigRuns{KeepMax: &keepMax},
	}

	svc := configsvc.New(&stubWorkspaceService{})
	written, err := svc.PutGlobal(context.Background(), doc)
	if err != nil {
		t.Fatalf("PutGlobal() error = %v", err)
	}
	if written.Runs == nil || written.Runs.KeepMax == nil || *written.Runs.KeepMax != keepMax {
		t.Fatalf("PutGlobal() written Runs.KeepMax = %v, want %d", written.Runs, keepMax)
	}

	reread, err := svc.GetGlobal(context.Background())
	if err != nil {
		t.Fatalf("GetGlobal() after PutGlobal error = %v", err)
	}
	if reread.Runs == nil || reread.Runs.KeepMax == nil || *reread.Runs.KeepMax != keepMax {
		t.Fatalf("GetGlobal() after PutGlobal Runs.KeepMax = %v, want %d", reread.Runs, keepMax)
	}
}

// TestPutWorkspaceConfigPreservesTaskRuntimeRules asserts that task_runtime_rules
// survive a PUT round-trip without being silently erased.
//
// The risk: ConfigTaskRun previously omitted the TaskRuntimeRules field, so any
// GUI save would reconstruct the config from a document that cannot represent the
// rules — permanently deleting them on every write. This test seeds a config with
// runtime rules, submits an unrelated change via PutWorkspace, and asserts the
// rules are still present in the reloaded config.
func TestPutWorkspaceConfigPreservesTaskRuntimeRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rcDir := filepath.Join(dir, ".rc")
	if err := os.MkdirAll(rcDir, 0o700); err != nil {
		t.Fatalf("mkdir .rc: %v", err)
	}

	// Seed an on-disk config that includes task_runtime_rules.
	// "claude" is a valid IDE value per the agent driver catalog.
	ideVal := model.IDEClaude
	modelVal := "claude-opus-4"
	typeVal := "feature"
	rules := []model.TaskRuntimeRule{
		{Type: &typeVal, IDE: &ideVal, Model: &modelVal},
	}
	seedCfg := workspace.ProjectConfig{
		Tasks: workspace.TasksConfig{
			Run: workspace.TaskRunConfig{
				TaskRuntimeRules: &rules,
			},
		},
	}
	seedData, err := toml.Marshal(seedCfg)
	if err != nil {
		t.Fatalf("toml.Marshal seed: %v", err)
	}
	configPath := filepath.Join(rcDir, "config.toml")
	if err := os.WriteFile(configPath, seedData, 0o600); err != nil {
		t.Fatalf("WriteFile seed: %v", err)
	}

	// Read the current config to get a document that includes the rules.
	svc := configsvc.New(&stubWorkspaceService{rootDir: dir})
	doc, err := svc.GetWorkspace(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("GetWorkspace() error = %v", err)
	}

	// Mutate only an unrelated field (Runs.KeepMax) and save.
	keepMax := 99
	doc.Runs = &contract.ConfigRuns{KeepMax: &keepMax}
	if _, err := svc.PutWorkspace(context.Background(), "ws-1", doc); err != nil {
		t.Fatalf("PutWorkspace() error = %v", err)
	}

	// Re-read and assert the rules survived.
	reread, err := svc.GetWorkspace(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("GetWorkspace() after put error = %v", err)
	}
	if reread.Tasks == nil || reread.Tasks.Run == nil {
		t.Fatal("after PutWorkspace: tasks.run is nil — task_runtime_rules were erased")
	}
	got := reread.Tasks.Run.TaskRuntimeRules
	if got == nil || len(*got) != 1 {
		t.Fatalf("after PutWorkspace: task_runtime_rules len = %v, want 1 rule", got)
	}
	rule := (*got)[0]
	if rule.Type == nil || *rule.Type != typeVal {
		t.Errorf("rule.Type = %v, want %q", rule.Type, typeVal)
	}
	if rule.IDE == nil || *rule.IDE != ideVal {
		t.Errorf("rule.IDE = %v, want %q", rule.IDE, ideVal)
	}
	if rule.Model == nil || *rule.Model != modelVal {
		t.Errorf("rule.Model = %v, want %q", rule.Model, modelVal)
	}
}
