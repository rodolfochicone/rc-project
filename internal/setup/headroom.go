package setup

// HeadroomBinaryName is the headroom executable resolved on PATH.
const HeadroomBinaryName = "headroom"

const (
	// headroomPackage is the published Python package, with the [all] extra to
	// pull in every optional feature of the CLI.
	headroomPackage = "headroom-ai[all]"
	// headroomRepoURL is the headroom source repository, used for guidance.
	headroomRepoURL = "https://github.com/headroomlabs-ai/headroom"
)

// ResolveHeadroomInstall returns the environment-appropriate install command
// for headroom, a Python package. pipx is preferred because it installs CLI
// tools into isolated environments; pip3/pip are used as fallbacks. When no
// Python installer is found, the command is non-runnable and carries manual
// guidance.
func ResolveHeadroomInstall(hasPipx, hasPip3, hasPip bool) InstallCommand {
	switch {
	case hasPipx:
		return InstallCommand{
			Display:  `pipx install "` + headroomPackage + `"`,
			Name:     "pipx",
			Args:     []string{"install", headroomPackage},
			Runnable: true,
		}
	case hasPip3:
		return InstallCommand{
			Display:  `pip3 install "` + headroomPackage + `"`,
			Name:     "pip3",
			Args:     []string{"install", headroomPackage},
			Runnable: true,
		}
	case hasPip:
		return InstallCommand{
			Display:  `pip install "` + headroomPackage + `"`,
			Name:     "pip",
			Args:     []string{"install", headroomPackage},
			Runnable: true,
		}
	default:
		return InstallCommand{
			Display:  `install Python 3.10+ then run: pip install "` + headroomPackage + `"`,
			Runnable: false,
			Manual: "Install Python 3.10+ (or pipx) and run: pip install \"" + headroomPackage +
				"\" (see " + headroomRepoURL + ").",
		}
	}
}
