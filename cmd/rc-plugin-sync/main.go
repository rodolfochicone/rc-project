// Command rc-plugin-sync rewrites the version field in the Claude Code plugin
// manifests to a target release version. It is invoked by the OSS GoReleaser
// before-hook with the resolved release tag so the manifests track each
// release automatically.
package main

import (
	"fmt"
	"os"

	"github.com/rodolfochicone/rc-project/internal/pluginmanifest"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "" {
		fmt.Fprintln(os.Stderr, "usage: rc-plugin-sync <version>")
		os.Exit(2)
	}

	if err := pluginmanifest.Sync(".claude-plugin", os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "rc-plugin-sync: %v\n", err)
		os.Exit(1)
	}
}
