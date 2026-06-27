package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/rodolfochicone/rc-project"
	"github.com/rodolfochicone/rc-project/internal/charmtheme"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/update"
	"github.com/rodolfochicone/rc-project/internal/version"
	"github.com/spf13/cobra"
)

const (
	// updateResultWaitTimeout bounds how long we wait, after the command
	// finishes, to display the update banner on the current run.
	updateResultWaitTimeout = 1 * time.Second
	// updateCheckTimeout bounds the background update check so a cold or stale
	// cache run can finish its network request and persist the result, while
	// preventing a slow network from stalling the CLI indefinitely.
	updateCheckTimeout = 3 * time.Second
)

var (
	newMainCommand     = rc.NewCommand
	startUpdateCheckFn = startUpdateCheck
)

func main() {
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	cmd := newMainCommand()
	cmd.Version = version.String()

	updateDoneCh := make(chan struct{})
	close(updateDoneCh)
	var updateDone <-chan struct{} = updateDoneCh
	var updateResult <-chan *update.ReleaseInfo
	cancelUpdateCheck := func() {}
	if shouldStartUpdateCheck(os.Args[1:]) {
		updateResult, cancelUpdateCheck, updateDone = startUpdateCheckFn(ctx, version.Version)
	}
	// Cancel only on return so the background check can finish its network
	// request and persist the 24h cache even after the command completes.
	defer cancelUpdateCheck()

	executedCmd, err := cmd.ExecuteC()

	if release := waitForUpdateResult(updateResult); release != nil && shouldWriteUpdateNotification(executedCmd) {
		if writeErr := writeUpdateNotification(
			cmd.ErrOrStderr(),
			version.Version,
			release,
		); writeErr != nil &&
			err == nil {
			err = fmt.Errorf("write update notification: %w", writeErr)
		}
	}
	// Wait for the background check to settle so its result is persisted to the
	// state cache; bounded by the check's own timeout (updateCheckTimeout).
	<-updateDone

	if err != nil {
		return rc.ExitCode(err)
	}
	return 0
}

func shouldStartUpdateCheck(args []string) bool {
	for _, arg := range args {
		switch strings.TrimSpace(arg) {
		case "-h", "--help", "--version":
			return false
		}
	}

	sawCommand := false
	for _, arg := range args {
		value := strings.TrimSpace(arg)
		if value == "" || strings.HasPrefix(value, "-") {
			continue
		}
		sawCommand = true
		switch value {
		case "help", "version", "completion", "__complete", "__completeNoDesc":
			return false
		}
	}

	return sawCommand
}

func startUpdateCheck(
	parent context.Context,
	currentVersion string,
) (<-chan *update.ReleaseInfo, context.CancelFunc, <-chan struct{}) {
	ctx, cancel := context.WithTimeout(parent, updateCheckTimeout)
	result := make(chan *update.ReleaseInfo, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(result)

		statePath, err := update.StateFilePath()
		if err != nil {
			return
		}

		release, err := update.CheckForUpdate(ctx, currentVersion, statePath)
		if err != nil || release == nil {
			return
		}

		result <- release
	}()

	return result, cancel, done
}

func waitForUpdateResult(result <-chan *update.ReleaseInfo) *update.ReleaseInfo {
	if result == nil {
		return nil
	}
	select {
	case release, ok := <-result:
		if !ok {
			return nil
		}
		return release
	case <-time.After(updateResultWaitTimeout):
		return nil
	}
}

func renderUpdateNotification(currentVersion string, release *update.ReleaseInfo) string {
	styles := updateNotificationStyles{
		header:  lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorWarning),
		current: lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorMuted),
		arrow:   lipgloss.NewStyle().Foreground(charmtheme.ColorMuted),
		latest:  lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorBrand),
		body:    lipgloss.NewStyle().Foreground(charmtheme.ColorMuted),
	}

	lineOne := fmt.Sprintf(
		"%s %s %s %s",
		styles.header.Render("Update available:"),
		styles.current.Render(strings.TrimSpace(currentVersion)),
		styles.arrow.Render("->"),
		styles.latest.Render(release.Version),
	)
	lineTwo := styles.body.Render("Run 'rc upgrade' to update")

	return lipgloss.JoinVertical(lipgloss.Left, lineOne, lineTwo)
}

func writeUpdateNotification(w io.Writer, currentVersion string, release *update.ReleaseInfo) error {
	if release == nil {
		return nil
	}
	_, err := fmt.Fprintln(w, renderUpdateNotification(currentVersion, release))
	return err
}

func shouldWriteUpdateNotification(cmd *cobra.Command) bool {
	if cmd == nil {
		return true
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		return true
	}

	switch strings.TrimSpace(formatFlag.Value.String()) {
	case model.OutputFormatJSONValue, model.OutputFormatRawJSONValue:
		return false
	default:
		return true
	}
}

type updateNotificationStyles struct {
	header  lipgloss.Style
	current lipgloss.Style
	arrow   lipgloss.Style
	latest  lipgloss.Style
	body    lipgloss.Style
}
