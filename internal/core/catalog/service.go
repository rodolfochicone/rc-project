// Package catalog implements CatalogService, a thin read-only adapter that
// surfaces installed extensions and reusable agents via the daemon HTTP API.
package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/rodolfochicone/rc-project/internal/api/contract"
	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/core/agents"
	extensions "github.com/rodolfochicone/rc-project/internal/core/extension"
)

// ExtensionDiscovery abstracts extension.Discovery for injection in tests.
type ExtensionDiscovery interface {
	Discover(ctx context.Context) (extensions.DiscoveryResult, error)
}

// workspaceLookup resolves a workspace root dir from an ID.
type workspaceLookup interface {
	Get(ctx context.Context, id string) (apicore.Workspace, error)
}

// Service implements apicore.CatalogService using the extension and agent
// discovery subsystems. Discovery objects are constructed per-call so they
// always reflect the current on-disk state.
type Service struct {
	workspaces    workspaceLookup
	newDiscovery  func(workspaceRoot string) ExtensionDiscovery
	agentRegistry AgentRegistry
}

var _ apicore.CatalogService = (*Service)(nil)

// Option is a functional option for Service.
type Option func(*Service)

// WithDiscoveryFactory replaces the production extension discovery factory.
// Intended for tests that need to inject a controlled discovery result.
func WithDiscoveryFactory(fn func(workspaceRoot string) ExtensionDiscovery) Option {
	return func(s *Service) { s.newDiscovery = fn }
}

// AgentRegistry is the interface for agent catalog discovery, exported so tests
// outside the package can inject a stub via WithAgentRegistry.
type AgentRegistry interface {
	Discover(ctx context.Context, workspaceRoot string) (agents.Catalog, error)
}

// WithAgentRegistry replaces the production agent registry.
// Intended for tests that need to inject a controlled catalog.
func WithAgentRegistry(r AgentRegistry) Option {
	return func(s *Service) { s.agentRegistry = r }
}

// New constructs a Service with the production extension discovery and agent
// registry implementations. Optional opts allow test injection.
func New(workspaces apicore.WorkspaceService, opts ...Option) *Service {
	s := &Service{
		workspaces: workspaces,
		newDiscovery: func(workspaceRoot string) ExtensionDiscovery {
			return extensions.Discovery{WorkspaceRoot: workspaceRoot}
		},
		agentRegistry: agents.New(),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Extensions discovers all extensions for the workspace identified by workspaceID
// and returns a read-only listing. Non-fatal discovery failures are surfaced as
// metadata counts, not errors, so the caller always receives whatever was found.
func (s *Service) Extensions(ctx context.Context, workspaceID string) (contract.ExtensionListResponse, error) {
	root, err := s.workspaceRootFor(ctx, workspaceID)
	if err != nil {
		return contract.ExtensionListResponse{}, err
	}

	result, err := s.newDiscovery(root).Discover(ctx)
	if err != nil {
		return contract.ExtensionListResponse{}, fmt.Errorf("discover extensions: %w", err)
	}

	items := make([]contract.ExtensionItem, 0, len(result.Extensions))
	for idx := range result.Extensions {
		ext := &result.Extensions[idx]
		item := contract.ExtensionItem{
			Name:    ext.Ref.Name,
			Source:  string(ext.Ref.Source),
			Enabled: ext.Enabled,
		}
		if ext.Manifest != nil {
			item.Description = ext.Manifest.Extension.Description
			item.Version = ext.Manifest.Extension.Version
		}
		items = append(items, item)
	}

	return contract.ExtensionListResponse{
		Extensions:    items,
		FailureCount:  len(result.Failures),
		OverrideCount: len(result.Overrides),
	}, nil
}

// Agents discovers all reusable agents for the workspace identified by workspaceID.
// Per-agent problems are surfaced as per-item warnings so the listing is always
// returned even when some agents fail to load.
func (s *Service) Agents(ctx context.Context, workspaceID string) (contract.AgentListResponse, error) {
	root, err := s.workspaceRootFor(ctx, workspaceID)
	if err != nil {
		return contract.AgentListResponse{}, err
	}

	catalog, err := s.agentRegistry.Discover(ctx, root)
	if err != nil {
		return contract.AgentListResponse{}, fmt.Errorf("discover agents: %w", err)
	}

	items := make([]contract.AgentItem, 0, len(catalog.Agents))
	for idx := range catalog.Agents {
		a := &catalog.Agents[idx]
		items = append(items, contract.AgentItem{
			Name:        a.Name,
			Scope:       string(a.Source.Scope),
			Description: a.Metadata.Description,
		})
	}

	// Append per-agent problems as stub items with a sanitized warning so the
	// UI can show the user which agents failed to load and why, without
	// exposing internal filesystem paths present in raw error strings.
	for idx := range catalog.Problems {
		p := &catalog.Problems[idx]
		items = append(items, contract.AgentItem{
			Name:     p.Name,
			Scope:    string(p.Source.Scope),
			Warnings: []string{sanitizeAgentLoadError(p.Err)},
		})
	}

	return contract.AgentListResponse{Agents: items}, nil
}

func (s *Service) workspaceRootFor(ctx context.Context, workspaceID string) (string, error) {
	ws, err := s.workspaces.Get(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("get workspace: %w", err)
	}
	return ws.RootDir, nil
}

// sanitizeAgentLoadError converts a raw agent load error into a message safe
// for API responses. Raw errors can embed absolute on-disk paths (e.g. the
// full path to a missing AGENT.md). We map known sentinel errors to clean
// descriptions and fall back to a generic message for unknown errors so no
// internal filesystem layout leaks to the caller.
func sanitizeAgentLoadError(err error) string {
	switch {
	case errors.Is(err, agents.ErrInvalidAgentName):
		return "invalid agent name"
	case errors.Is(err, agents.ErrReservedAgentName):
		return "reserved agent name"
	case errors.Is(err, agents.ErrMissingAgentDefinition):
		return "agent definition file (AGENT.md) is missing"
	case errors.Is(err, agents.ErrMalformedFrontmatter):
		return "agent definition has malformed front matter"
	case errors.Is(err, agents.ErrUnsupportedMetadataField):
		return "agent definition contains an unsupported metadata field"
	case errors.Is(err, agents.ErrInvalidRuntimeDefaults):
		return "agent has invalid runtime defaults"
	case errors.Is(err, agents.ErrMalformedMCPConfig):
		return "agent mcp.json is malformed"
	case errors.Is(err, agents.ErrMissingEnvironmentVariable):
		return "agent references a missing environment variable"
	default:
		return "failed to load agent"
	}
}
