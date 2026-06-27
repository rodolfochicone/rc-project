package core_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rodolfochicone/rc-project/internal/api/contract"
	"github.com/rodolfochicone/rc-project/internal/api/core"
)

// TestHandlerConfigIncludesConfigAndCatalogServices asserts that HandlerConfig
// carries ConfigService and CatalogService fields. This is the structural gate
// for T6: the fields must exist before any wiring or handler can use them.
// If the fields are absent, this file will not compile — compilation failure
// is the intended red state.
func TestHandlerConfigIncludesConfigAndCatalogServices(t *testing.T) {
	t.Parallel()

	cfg := core.HandlerConfig{
		Config:  nil,
		Catalog: nil,
	}
	_ = cfg
}

// TestHandlersCarryConfigAndCatalogFields asserts that NewHandlers accepts and
// stores Config and Catalog so handlers can invoke them.
func TestHandlersCarryConfigAndCatalogFields(t *testing.T) {
	t.Parallel()

	h := core.NewHandlers(&core.HandlerConfig{
		Config:  nil,
		Catalog: nil,
	})
	_ = h
}

// TestConfigHandlerGlobalGetServiceUnavailable asserts that GET /api/config/global
// returns 503 service_unavailable when no ConfigService is wired. This mirrors
// the existing pattern for every other service in the handlers error-path tests,
// and ensures the route is registered and guarded consistently.
func TestConfigHandlerGlobalGetServiceUnavailable(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handlers := core.NewHandlers(&core.HandlerConfig{TransportName: "test"})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/config/global",
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var payload core.TransportError
	decodeJSON(t, rec.Body.Bytes(), &payload)
	if payload.Code != "service_unavailable" {
		t.Fatalf("code = %q, want service_unavailable", payload.Code)
	}
}

// TestConfigHandlerGlobalPutServiceUnavailable asserts that PUT /api/config/global
// returns 503 when no ConfigService is wired.
func TestConfigHandlerGlobalPutServiceUnavailable(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handlers := core.NewHandlers(&core.HandlerConfig{TransportName: "test"})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		"/api/config/global",
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var payload core.TransportError
	decodeJSON(t, rec.Body.Bytes(), &payload)
	if payload.Code != "service_unavailable" {
		t.Fatalf("code = %q, want service_unavailable", payload.Code)
	}
}

// TestConfigHandlerWorkspaceGetMissingWorkspaceContext asserts that
// GET /api/config/workspace returns 412 workspace_context_missing when the
// X-rc-Workspace-ID header is absent. This is the workspace-context guard
// required by T11: workspace-scoped config routes must enforce the header.
func TestConfigHandlerWorkspaceGetMissingWorkspaceContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handlers := core.NewHandlers(&core.HandlerConfig{
		TransportName: "test",
		Config:        &fakeConfigService{},
		Workspaces:    &smokeWorkspaceService{},
	})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/config/workspace",
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("status = %d, want %d (workspace_context_missing)",
			rec.Code, http.StatusPreconditionFailed)
	}
	var payload core.TransportError
	decodeJSON(t, rec.Body.Bytes(), &payload)
	if payload.Code != "workspace_context_missing" {
		t.Fatalf("code = %q, want workspace_context_missing", payload.Code)
	}
}

// TestCatalogHandlerExtensionsMissingWorkspaceContext asserts that
// GET /api/catalog/extensions returns 412 when the workspace header is absent.
func TestCatalogHandlerExtensionsMissingWorkspaceContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handlers := core.NewHandlers(&core.HandlerConfig{
		TransportName: "test",
		Catalog:       &fakeCatalogService{},
		Workspaces:    &smokeWorkspaceService{},
	})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/catalog/extensions",
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("status = %d, want %d (workspace_context_missing)",
			rec.Code, http.StatusPreconditionFailed)
	}
	var payload core.TransportError
	decodeJSON(t, rec.Body.Bytes(), &payload)
	if payload.Code != "workspace_context_missing" {
		t.Fatalf("code = %q, want workspace_context_missing", payload.Code)
	}
}

// TestCatalogHandlerAgentsMissingWorkspaceContext asserts that
// GET /api/catalog/agents returns 412 when the workspace header is absent.
func TestCatalogHandlerAgentsMissingWorkspaceContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handlers := core.NewHandlers(&core.HandlerConfig{
		TransportName: "test",
		Catalog:       &fakeCatalogService{},
		Workspaces:    &smokeWorkspaceService{},
	})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/catalog/agents",
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("status = %d, want %d (workspace_context_missing)",
			rec.Code, http.StatusPreconditionFailed)
	}
	var payload core.TransportError
	decodeJSON(t, rec.Body.Bytes(), &payload)
	if payload.Code != "workspace_context_missing" {
		t.Fatalf("code = %q, want workspace_context_missing", payload.Code)
	}
}

// TestGlobalConfigPathDoesNotRequireWorkspaceHeader asserts that
// GET /api/config/global does NOT enforce the workspace header, because
// global config is not workspace-scoped. This prevents the T11 guard from
// being over-applied.
func TestGlobalConfigPathDoesNotRequireWorkspaceHeader(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handlers := core.NewHandlers(&core.HandlerConfig{
		TransportName: "test",
		Config:        &fakeConfigService{},
		Workspaces:    &smokeWorkspaceService{},
	})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/config/global",
		http.NoBody,
	)
	// No X-rc-Workspace-ID header — must NOT get 412.
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code == http.StatusPreconditionFailed {
		t.Fatal("GET /api/config/global returned 412 without workspace header — " +
			"global config must not require workspace context")
	}
}

// TestHandlerClonePreservesConfigAndCatalogServices asserts that Clone() copies
// the Config and Catalog fields into the cloned Handlers, so each transport
// gets its own handler set without losing the wired services.
func TestHandlerClonePreservesConfigAndCatalogServices(t *testing.T) {
	t.Parallel()

	configSvc := &fakeConfigService{}
	catalogSvc := &fakeCatalogService{}

	h := core.NewHandlers(&core.HandlerConfig{
		TransportName: "test",
		Config:        configSvc,
		Catalog:       catalogSvc,
	})
	cloned := h.Clone()

	if cloned.Config != configSvc {
		t.Fatal("Clone() did not preserve Config service")
	}
	if cloned.Catalog != catalogSvc {
		t.Fatal("Clone() did not preserve Catalog service")
	}
}

// fakeConfigService is a minimal ConfigService stub for tests that need the
// interface wired but don't test its behavior.
// The interface signature will be defined in core.ConfigService (T6).
type fakeConfigService struct{}

var _ core.ConfigService = (*fakeConfigService)(nil)

func (*fakeConfigService) GetGlobal(_ context.Context) (contract.ConfigDocument, error) {
	return contract.ConfigDocument{}, nil
}

func (*fakeConfigService) PutGlobal(_ context.Context, doc contract.ConfigDocument) (contract.ConfigDocument, error) {
	return doc, nil
}

func (*fakeConfigService) GetWorkspace(_ context.Context, _ string) (contract.ConfigDocument, error) {
	return contract.ConfigDocument{}, nil
}

func (*fakeConfigService) PutWorkspace(
	_ context.Context,
	_ string,
	doc contract.ConfigDocument,
) (contract.ConfigDocument, error) {
	return doc, nil
}

// fakeCatalogService is a minimal CatalogService stub.
// The interface signature will be defined in core.CatalogService (T6).
type fakeCatalogService struct{}

var _ core.CatalogService = (*fakeCatalogService)(nil)

func (*fakeCatalogService) Extensions(_ context.Context, _ string) (contract.ExtensionListResponse, error) {
	return contract.ExtensionListResponse{}, nil
}

func (*fakeCatalogService) Agents(_ context.Context, _ string) (contract.AgentListResponse, error) {
	return contract.AgentListResponse{}, nil
}
