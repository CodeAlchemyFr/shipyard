package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "shipyard",
	Short: "Shipyard CLI - Deploy applications to Kubernetes with ease",
	Long: `Shipyard is a PaaS CLI tool that simplifies Kubernetes deployments.
It generates Kubernetes manifests and manages deployments for your applications.`,
	Version: version,
}

// SetVersion sets the version for the CLI
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(releasesCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(domainCmd)
	rootCmd.AddCommand(registryCmd)
}