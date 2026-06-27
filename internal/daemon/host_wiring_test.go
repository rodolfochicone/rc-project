package daemon

import (
	"testing"

	"github.com/rodolfochicone/rc-project/internal/core/catalog"
	"github.com/rodolfochicone/rc-project/internal/core/configsvc"
)

// TestBuildHostHandlersWiresConfigAndCatalog asserts that buildHostHandlers
// assigns non-nil Config and Catalog services to the handler set.
//
// This is the regression guard for the wiring gap: when Config or Catalog are
// nil, every corresponding handler returns 503 service_unavailable, making the
// entire config/catalog backend non-functional despite the routes being
// registered. The previous pipeline was green because handler tests injected
// fake services directly, bypassing this assembly path entirely.
func TestBuildHostHandlersWiresConfigAndCatalog(t *testing.T) {
	t.Parallel()

	env := newRunManagerTestEnv(t, runManagerTestDeps{})

	persistence := hostPersistence{
		db:              env.globalDB,
		settings:        RunLifecycleSettings{},
		reconcileResult: ReconcileResult{},
	}

	handlers := buildHostHandlers(
		&Host{paths: env.paths},
		persistence,
		env.manager,
		func() {},
	)

	if handlers.Config == nil {
		t.Fatal("buildHostHandlers() produced nil Config — all config endpoints return 503")
	}
	if handlers.Catalog == nil {
		t.Fatal("buildHostHandlers() produced nil Catalog — all catalog endpoints return 503")
	}

	// Verify the concrete types so a future type-mismatch (e.g., wrong service
	// passed to the wrong field) is caught at this level.
	if _, ok := handlers.Config.(*configsvc.Service); !ok {
		t.Fatalf("handlers.Config type = %T, want *configsvc.Service", handlers.Config)
	}
	if _, ok := handlers.Catalog.(*catalog.Service); !ok {
		t.Fatalf("handlers.Catalog type = %T, want *catalog.Service", handlers.Catalog)
	}
}
