package setup

// RTKBinaryName is the rtk executable resolved on PATH.
const RTKBinaryName = "rtk"

// runtime.GOOS values branched on when resolving installers.
const (
	goosDarwin  = "darwin"
	goosWindows = "windows"
)

const (
	// rtkInstallScriptURL is the official cross-platform (macOS/Linux) install script.
	rtkInstallScriptURL = "https://raw.githubusercontent.com/rtk-ai/rtk/refs/heads/master/install.sh"
	// rtkRepoURL is the rtk source repository, used for cargo installs and guidance.
	rtkRepoURL = "https://github.com/rtk-ai/rtk"
	// rtkReleasesURL points users to prebuilt binaries when no installer can run.
	rtkReleasesURL = "https://github.com/rtk-ai/rtk/releases"
)

// ResolveRTKInstall returns the OS-appropriate install command for rtk. goos is
// a runtime.GOOS value; hasBrew and hasCargo report whether those package
// managers were found on PATH.
func ResolveRTKInstall(goos string, hasBrew, hasCargo bool) InstallCommand {
	switch goos {
	case goosDarwin:
		if hasBrew {
			return rtkBrewCommand()
		}
		return rtkScriptCommand()
	case goosWindows:
		if hasCargo {
			return rtkCargoCommand()
		}
		return InstallCommand{
			Display:  "download a prebuilt binary from " + rtkReleasesURL,
			Runnable: false,
			Manual:   "Download the Windows binary from " + rtkReleasesURL + " and add it to PATH.",
		}
	default:
		// Linux and other Unix-like systems use the official install script.
		return rtkScriptCommand()
	}
}

func rtkBrewCommand() InstallCommand {
	return InstallCommand{
		Display:  "brew install rtk",
		Name:     "brew",
		Args:     []string{"install", "rtk"},
		Runnable: true,
	}
}

func rtkCargoCommand() InstallCommand {
	return InstallCommand{
		Display:  "cargo install --git " + rtkRepoURL,
		Name:     "cargo",
		Args:     []string{"install", "--git", rtkRepoURL},
		Runnable: true,
	}
}

func rtkScriptCommand() InstallCommand {
	script := "curl -fsSL " + rtkInstallScriptURL + " | sh"
	return InstallCommand{
		Display:  script,
		Name:     "sh",
		Args:     []string{"-c", script},
		Runnable: true,
	}
}
