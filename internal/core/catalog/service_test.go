package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/core/agents"
	"github.com/rodolfochicone/rc-project/internal/core/catalog"
	extensions "github.com/rodolfochicone/rc-project/internal/core/extension"
)

// TestServiceImplementsCatalogService is a compile-time assertion.
// It fails to compile if *catalog.Service no longer satisfies the interface.
var _ apicore.CatalogService = (*catalog.Service)(nil)

// TestExtensionsReturnsEmptyListOnNoExtensions asserts that Extensions returns
// a non-nil, zero-length slice when no extensions are installed. This matters
// because the UI must distinguish "no extensions" from "discovery failed".
func TestExtensionsReturnsEmptyListOnNoExtensions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	svc := catalog.New(&stubWorkspaceService{rootDir: dir})

	result, err := svc.Extensions(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("Extensions() error = %v", err)
	}
	if result.Extensions == nil {
		t.Fatal("Extensions() returned nil slice, want non-nil empty slice")
	}
}

// TestAgentsReturnsEmptyListOnNoAgents asserts that Agents returns a non-nil,
// zero-length slice when no agents are installed. This matters for the same UI
// reason as the extension case above.
func TestAgentsReturnsEmptyListOnNoAgents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	svc := catalog.New(&stubWorkspaceService{rootDir: dir})

	result, err := svc.Agents(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("Agents() error = %v", err)
	}
	if result.Agents == nil {
		t.Fatal("Agents() returned nil slice, want non-nil empty slice")
	}
}

// TestExtensionsWorkspaceServiceErrorPropagates asserts that a workspace lookup
// failure is returned as an error rather than swallowed. This matters because
// silently returning an empty list would hide misconfiguration from the user.
func TestExtensionsWorkspaceServiceErrorPropagates(t *testing.T) {
	t.Parallel()

	boom := errors.New("workspace store offline")
	svc := catalog.New(&stubWorkspaceService{err: boom})

	_, err := svc.Extensions(context.Background(), "ws-1")
	if err == nil {
		t.Fatal("Extensions() with workspace error: want error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("Extensions() error = %v, want to wrap %v", err, boom)
	}
}

// TestAgentsWorkspaceServiceErrorPropagates mirrors the extension test for Agents.
func TestAgentsWorkspaceServiceErrorPropagates(t *testing.T) {
	t.Parallel()

	boom := errors.New("workspace store offline")
	svc := catalog.New(&stubWorkspaceService{err: boom})

	_, err := svc.Agents(context.Background(), "ws-1")
	if err == nil {
		t.Fatal("Agents() with workspace error: want error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("Agents() error = %v, want to wrap %v", err, boom)
	}
}

// stubWorkspaceService satisfies apicore.WorkspaceService for catalog tests.
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
func (s *stubWorkspaceService) List(_ context.Context) ([]apicore.Workspace, error) { return nil, nil }

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

// TestExtensionsItemShape asserts that ExtensionItem fields are populated from
// the discovered extension manifest. This matters because the UI depends on
// name, source, description, and version to render each extension row — missing
// any field silently produces blank cells in the table.
func TestExtensionsItemShape(t *testing.T) {
	t.Parallel()

	wantName := "my-ext"
	wantSource := string(extensions.SourceUser)
	wantDescription := "does things"
	wantVersion := "1.2.3"

	fakeResult := extensions.DiscoveryResult{
		Extensions: []extensions.DiscoveredExtension{
			{
				Ref:     extensions.Ref{Name: wantName, Source: extensions.SourceUser},
				Enabled: true,
				Manifest: &extensions.Manifest{
					Extension: extensions.ExtensionInfo{
						Name:        wantName,
						Version:     wantVersion,
						Description: wantDescription,
					},
				},
			},
		},
	}

	dir := t.TempDir()
	svc := catalog.New(
		&stubWorkspaceService{rootDir: dir},
		catalog.WithDiscoveryFactory(func(_ string) catalog.ExtensionDiscovery {
			return &stubExtensionDiscovery{result: fakeResult}
		}),
	)

	result, err := svc.Extensions(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("Extensions() error = %v", err)
	}
	if len(result.Extensions) != 1 {
		t.Fatalf("Extensions() len = %d, want 1", len(result.Extensions))
	}

	item := result.Extensions[0]
	if item.Name != wantName {
		t.Errorf("item.Name = %q, want %q", item.Name, wantName)
	}
	if item.Source != wantSource {
		t.Errorf("item.Source = %q, want %q", item.Source, wantSource)
	}
	if item.Description != wantDescription {
		t.Errorf("item.Description = %q, want %q", item.Description, wantDescription)
	}
	if item.Version != wantVersion {
		t.Errorf("item.Version = %q, want %q", item.Version, wantVersion)
	}
	if !item.Enabled {
		t.Errorf("item.Enabled = false, want true")
	}
}

// stubExtensionDiscovery satisfies catalog.ExtensionDiscovery for injection in tests.
type stubExtensionDiscovery struct {
	result extensions.DiscoveryResult
	err    error
}

func (s *stubExtensionDiscovery) Discover(_ context.Context) (extensions.DiscoveryResult, error) {
	return s.result, s.err
}

// stubAgentRegistry satisfies catalog.AgentRegistry for injection in tests.
type stubAgentRegistry struct {
	catalog agents.Catalog
	err     error
}

func (s *stubAgentRegistry) Discover(_ context.Context, _ string) (agents.Catalog, error) {
	return s.catalog, s.err
}

// TestAgentWarningsSanitized asserts that raw error strings containing
// filesystem paths are never surfaced in the API response. The daemon is
// localhost-only but any process on loopback could read workspace layout from
// an unsanitized error. Known sentinel errors must produce a clean description;
// unknown errors must produce a generic message with no internal detail.
func TestAgentWarningsSanitized(t *testing.T) {
	t.Parallel()

	agentPath := "/home/user/.rc/agents/bad-agent/AGENT.md"

	cases := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "missing definition with path",
			err:     fmt.Errorf("%w: %s", agents.ErrMissingAgentDefinition, agentPath),
			wantMsg: "agent definition file (AGENT.md) is missing",
		},
		{
			name:    "malformed frontmatter",
			err:     agents.ErrMalformedFrontmatter,
			wantMsg: "agent definition has malformed front matter",
		},
		{
			name:    "unsupported metadata field",
			err:     agents.ErrUnsupportedMetadataField,
			wantMsg: "agent definition contains an unsupported metadata field",
		},
		{
			name:    "unknown error hides detail",
			err:     fmt.Errorf("read %s: permission denied", agentPath),
			wantMsg: "failed to load agent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			problemCatalog := agents.Catalog{
				Agents: []agents.ResolvedAgent{},
				Problems: []agents.Problem{
					{Name: "bad-agent", Err: tc.err},
				},
			}
			svc := catalog.New(
				&stubWorkspaceService{rootDir: dir},
				catalog.WithAgentRegistry(&stubAgentRegistry{catalog: problemCatalog}),
			)

			result, err := svc.Agents(context.Background(), "ws-1")
			if err != nil {
				t.Fatalf("Agents() error = %v", err)
			}
			if len(result.Agents) != 1 {
				t.Fatalf("Agents() len = %d, want 1 (the problem item)", len(result.Agents))
			}
			item := result.Agents[0]
			if len(item.Warnings) != 1 {
				t.Fatalf("item.Warnings len = %d, want 1", len(item.Warnings))
			}
			got := item.Warnings[0]
			if got != tc.wantMsg {
				t.Errorf("warning = %q, want %q", got, tc.wantMsg)
			}
			if got == tc.err.Error() {
				t.Errorf("warning equals raw error string — filesystem paths may be exposed: %q", got)
			}
		})
	}
}
