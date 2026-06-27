package contract_test

import (
	"net/http"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/api/contract"
)

// TestConfigCatalogRouteInventoryEntries asserts that the six new config/catalog
// routes are present in the RouteInventory with the correct HTTP methods and
// timeout classes. This matters because openapi_contract_test.go enforces
// inventory↔runtime parity, so missing entries here cause both the contract and
// integration tests to fail.
func TestConfigCatalogRouteInventoryEntries(t *testing.T) {
	t.Parallel()

	wantRoutes := []struct {
		method       string
		path         string
		responseType string
		timeoutClass contract.TimeoutClass
	}{
		{http.MethodGet, "/api/config/global", "ConfigDocumentResponse", contract.TimeoutRead},
		{http.MethodPut, "/api/config/global", "ConfigDocumentResponse", contract.TimeoutMutate},
		{http.MethodGet, "/api/config/workspace", "ConfigDocumentResponse", contract.TimeoutRead},
		{http.MethodPut, "/api/config/workspace", "ConfigDocumentResponse", contract.TimeoutMutate},
		{http.MethodGet, "/api/catalog/extensions", "ExtensionListResponse", contract.TimeoutRead},
		{http.MethodGet, "/api/catalog/agents", "AgentListResponse", contract.TimeoutRead},
	}

	for _, want := range wantRoutes {
		t.Run(want.method+" "+want.path, func(t *testing.T) {
			t.Parallel()

			route, ok := contract.FindRoute(want.method, want.path)
			if !ok {
				t.Fatalf("route %s %s not found in RouteInventory", want.method, want.path)
			}
			if route.TimeoutClass != want.timeoutClass {
				t.Fatalf("TimeoutClass for %s %s = %v, want %v",
					want.method, want.path, route.TimeoutClass, want.timeoutClass)
			}
			if route.ResponseType != want.responseType {
				t.Fatalf("ResponseType for %s %s = %q, want %q",
					want.method, want.path, route.ResponseType, want.responseType)
			}
		})
	}
}
