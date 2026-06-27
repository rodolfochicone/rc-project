package providerdefaults

import (
	"github.com/rodolfochicone/rc-project/internal/core/provider"
	"github.com/rodolfochicone/rc-project/internal/core/provider/coderabbit"
)

func DefaultRegistry() *provider.Registry {
	return DefaultRegistryForWorkspace("")
}

func DefaultRegistryForWorkspace(workspaceRoot string) *provider.Registry {
	registry := provider.NewRegistry()
	registry.Register(coderabbit.New(coderabbit.WithWorkingDir(workspaceRoot)))
	return registry
}
