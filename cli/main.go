package main

import (
	"fmt"
	"os"

	"github.com/shipyard/cli/cmd"
	versionpkg "github.com/shipyard/cli/pkg/version"
)

var version = "dev"

func main() {
	// Set version in both places
	cmd.SetVersion(version)
	versionpkg.Current = version
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}