package cli

import (
	"fmt"

	"github.com/rodolfochicone/rc-project/internal/update"
	"github.com/rodolfochicone/rc-project/internal/version"
	"github.com/spf13/cobra"
)

func newUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "upgrade",
		Aliases:      []string{"update"},
		Short:        "Upgrade rc to the latest release",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Upgrade rc using the appropriate installation flow for this machine.

Package-manager installs print the correct command to run. Direct binary installs
perform an in-place self-update.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signalCommandContext(cmd)
			defer stop()

			if err := update.Upgrade(ctx, version.Version, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("upgrade rc: %w", err)
			}
			return nil
		},
	}

	return cmd
}
