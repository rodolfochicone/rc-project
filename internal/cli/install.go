package cli

import (
	"io"

	"charm.land/lipgloss/v2"
	"github.com/rodolfochicone/rc-project/internal/core/kernel"
	"github.com/spf13/cobra"
)

// installResource is one resource installable via `rc install --<flag>`.
type installResource struct {
	// flag is the boolean flag name that selects this resource (e.g. "rtk").
	flag string
	// help describes the resource in flag help and the listing output.
	help string
	// selected is bound to the resource's flag.
	selected bool
	// installer performs detection and installation for the resource.
	installer *toolInstaller
}

// installCommandState holds the flags and resource registry for `rc install`,
// a command that installs individual rc resources directly, without running
// the full `rc setup` flow.
type installCommandState struct {
	yes       bool
	guide     bool
	resources []*installResource
}

func newInstallCommandState() *installCommandState {
	return &installCommandState{
		resources: []*installResource{
			{flag: "rtk", help: "Install the rtk runtime toolkit", installer: newRTKInstaller()},
			{flag: "headroom", help: "Install the headroom AI toolkit", installer: newHeadroomInstaller()},
		},
	}
}

func newInstallCommand(_ *kernel.Dispatcher) *cobra.Command {
	state := newInstallCommandState()
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "Install an individual rc resource (e.g. rtk or headroom)",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Install a single rc resource selected by flag, without running the full
rc setup flow. Run "rc install" with no flag to list the installable resources,
or add --guide to print a resource's getting-started tutorial without installing.`,
		Example: `  rc install
  rc install --rtk
  rc install --headroom --yes
  rc install --rtk --guide`,
		RunE: state.run,
	}

	for _, r := range state.resources {
		cmd.Flags().BoolVar(&r.selected, r.flag, false, r.help)
	}
	cmd.Flags().BoolVarP(&state.yes, "yes", "y", false, "Skip confirmation prompts and install unattended")
	cmd.Flags().
		BoolVar(&state.guide, "guide", false, "Print the selected resource's getting-started tutorial without installing")
	return cmd
}

func (s *installCommandState) run(cmd *cobra.Command, _ []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	w := cmd.OutOrStdout()
	selected := s.selectedResources()
	if len(selected) == 0 {
		s.printResources(w)
		return nil
	}

	if s.guide {
		for _, r := range selected {
			r.installer.printGuide(w)
		}
		return nil
	}

	for _, r := range selected {
		r.installer.yes = s.yes
		r.installer.runUnattended = true
		if err := r.installer.ensure(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

func (s *installCommandState) selectedResources() []*installResource {
	selected := make([]*installResource, 0, len(s.resources))
	for _, r := range s.resources {
		if r.selected {
			selected = append(selected, r)
		}
	}
	return selected
}

// printResources lists every installable resource and the flag that selects it.
func (s *installCommandState) printResources(w io.Writer) {
	styles := newCLIChromeStyles()
	lipgloss.Fprintln(w, styles.sectionTitle.Render("Installable resources"))
	for _, r := range s.resources {
		lipgloss.Fprintf(
			w,
			"  %s  %s\n",
			styles.label.Render("--"+r.flag),
			styles.value.Render(r.help),
		)
	}
	lipgloss.Fprintf(w, "\n  %s\n", styles.path.Render("Select one, e.g. `rc install --rtk`."))
}
