package run

import (
	"context"

	"github.com/rodolfochicone/rc-project/internal/core/model"
	execpkg "github.com/rodolfochicone/rc-project/internal/core/run/exec"
	executorpkg "github.com/rodolfochicone/rc-project/internal/core/run/executor"
	"github.com/rodolfochicone/rc-project/internal/core/run/journal"
	"github.com/rodolfochicone/rc-project/pkg/rc/events"
)

var execute = executorpkg.Execute
var executeExec = execpkg.ExecuteExec

func Execute(
	ctx context.Context,
	jobs []model.Job,
	runArtifacts model.RunArtifacts,
	runJournal *journal.Journal,
	bus *events.Bus[events.Event],
	cfg *model.RuntimeConfig,
	manager model.RuntimeManager,
) error {
	return execute(ctx, jobs, runArtifacts, runJournal, bus, cfg, manager)
}

func ExecuteExec(ctx context.Context, cfg *model.RuntimeConfig, scope model.RunScope) error {
	return executeExec(ctx, cfg, scope)
}
