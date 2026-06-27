package core

import (
	"context"

	migrationpkg "github.com/rodolfochicone/rc-project/internal/core/migration"
)

func migrateArtifacts(ctx context.Context, cfg MigrationConfig) (*MigrationResult, error) {
	return migrationpkg.Migrate(ctx, cfg)
}
