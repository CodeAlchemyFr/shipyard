package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/manifests"
	"github.com/shipyard/cli/pkg/k8s"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [version|image-tag]",
	Short: "Rollback to a previous deployment version",
	Long: `Rollback your application to a previous deployment version.
You can specify either a version (e.g., v1634567890) or an image tag (e.g., v1.2.3).
If no version is specified, it will show an interactive list of deployments.`,
	Run: func(cmd *cobra.Command, args []string) {
		var targetVersion string
		if len(args) > 0 {
			targetVersion = args[0]
			if err := runRollback(targetVersion); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
		} else {
			// Interactive mode when no version specified
			if err := runRollbackInteractive(); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
		}
	},
}

func runRollback(targetIdentifier string) error {
	fmt.Println("🔄 Starting rollback...")

	// Parse current config to get app name
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// Load version manager
	versionManager := manifests.NewVersionManager(config.App.Name)

	var targetVersion *manifests.DeploymentVersion

	if targetIdentifier == "" {
		// No version specified, get latest successful
		fmt.Println("🔍 Finding latest successful deployment...")
		targetVersion, err = versionManager.GetLatestSuccessfulVersion()
		if err != nil {
			return fmt.Errorf("failed to find successful deployment: %w", err)
		}
		fmt.Printf("📍 Found latest successful: %s (%s)\n", targetVersion.Version, targetVersion.ImageTag)
	} else {
		// Specific version requested
		fmt.Printf("🔍 Looking for version: %s\n", targetIdentifier)
		targetVersion, err = versionManager.GetVersionByIdentifier(targetIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find version %s: %w", targetIdentifier, err)
		}
	}

	// Confirm rollback
	fmt.Printf("🎯 Rolling back to:\n")
	fmt.Printf("   Version: %s\n", targetVersion.Version)
	fmt.Printf("   Image: %s\n", targetVersion.Image)
	fmt.Printf("   Deployed: %s\n", targetVersion.Timestamp.Format("2006-01-02 15:04:05"))

	// Update current config with target version's image
	config.App.Image = targetVersion.Image

	// Create new deployment version for the rollback
	newVersionManager := manifests.NewVersionManager(config.App.Name)
	rollbackVersion, err := newVersionManager.GenerateVersion(config)
	if err != nil {
		return fmt.Errorf("failed to create rollback version: %w", err)
	}

	// Mark this as a rollback
	rollbackVersion.RollbackTo = targetVersion.Version

	// Save the rollback version
	if err := newVersionManager.SaveVersion(rollbackVersion); err != nil {
		return fmt.Errorf("failed to save rollback version: %w", err)
	}

	// Generate manifests with rollback version
	generator := manifests.NewGeneratorWithVersion(config, rollbackVersion)
	
	fmt.Printf("📦 Generating rollback manifests...\n")
	
	if err := generator.GenerateAppManifests(); err != nil {
		// Mark rollback as failed
		newVersionManager.UpdateVersionStatus(rollbackVersion.Version, "failed")
		return fmt.Errorf("failed to generate rollback manifests: %w", err)
	}

	// Update ingress if domains changed
	if err := generator.UpdateIngressManifests(); err != nil {
		// Mark rollback as failed
		newVersionManager.UpdateVersionStatus(rollbackVersion.Version, "failed")
		return fmt.Errorf("failed to update ingress: %w", err)
	}

	// Apply manifests to Kubernetes
	fmt.Println("☸️  Applying rollback to Kubernetes cluster...")
	client, err := k8s.NewClient()
	if err != nil {
		// Mark rollback as failed
		newVersionManager.UpdateVersionStatus(rollbackVersion.Version, "failed")
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	if err := client.ApplyManifests(config.App.Name); err != nil {
		// Mark rollback as failed
		newVersionManager.UpdateVersionStatus(rollbackVersion.Version, "failed")
		return fmt.Errorf("failed to apply rollback manifests: %w", err)
	}

	// Mark rollback as successful
	if err := newVersionManager.UpdateVersionStatus(rollbackVersion.Version, "success"); err != nil {
		fmt.Printf("⚠️  Warning: failed to update version status: %v\n", err)
	}

	fmt.Printf("✅ Rollback successful!\n")
	fmt.Printf("   Rolled back from current to %s (%s)\n", targetVersion.Version, targetVersion.ImageTag)
	fmt.Printf("   New deployment version: %s\n", rollbackVersion.Version)

	return nil
}

// runRollbackInteractive provides an interactive rollback menu
func runRollbackInteractive() error {
	fmt.Println("🔄 Interactive Rollback")
	fmt.Println("======================")

	// Parse current config to get app name
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	fmt.Printf("App: %s\n", config.App.Name)

	// Load version manager
	versionManager := manifests.NewVersionManager(config.App.Name)

	// Get deployment history
	versions, err := versionManager.GetVersionHistory(10) // Get last 10 versions
	if err != nil {
		return fmt.Errorf("failed to get deployment history: %w", err)
	}

	if len(versions) == 0 {
		fmt.Println("📋 No deployment history found")
		return nil
	}

	// Display deployment options
	fmt.Println("\nAvailable deployments:")
	fmt.Printf("%-5s %-15s %-25s %-10s %-20s\n", "No.", "Version", "Image Tag", "Status", "Deployed")
	fmt.Println("─────────────────────────────────────────────────────────────────────────")

	successfulVersions := []manifests.DeploymentVersion{}
	for i, version := range versions {
		// Only show successful deployments for rollback
		if version.Status == "success" {
			successfulVersions = append(successfulVersions, version)
			
			statusIcon := "✅"
			if version.Status == "failed" {
				statusIcon = "❌"
			} else if version.Status == "pending" {
				statusIcon = "⏳"
			}

			fmt.Printf("%-5d %-15s %-25s %-10s %-20s\n", 
				len(successfulVersions), 
				version.Version, 
				version.ImageTag,
				statusIcon,
				version.Timestamp.Format("2006-01-02 15:04"),
			)
		}
	}

	if len(successfulVersions) == 0 {
		fmt.Println("📋 No successful deployments found for rollback")
		return nil
	}

	fmt.Println("  0. Cancel")

	fmt.Print("\nSelect deployment to rollback to: ")
	var choice string
	fmt.Scanln(&choice)

	if choice == "0" || strings.TrimSpace(choice) == "" {
		fmt.Println("❌ Rollback cancelled")
		return nil
	}

	index, err := strconv.Atoi(strings.TrimSpace(choice))
	if err != nil || index < 1 || index > len(successfulVersions) {
		return fmt.Errorf("invalid selection: must be between 1-%d", len(successfulVersions))
	}

	selectedVersion := successfulVersions[index-1]

	// Show rollback confirmation
	fmt.Printf("\n🎯 Rollback Details:\n")
	fmt.Printf("   From: %s (%s)\n", config.App.Image, "current")
	fmt.Printf("   To: %s (%s)\n", selectedVersion.Image, selectedVersion.Version)
	fmt.Printf("   Deployed: %s\n", selectedVersion.Timestamp.Format("2006-01-02 15:04:05"))
	
	// Confirm rollback
	fmt.Print("\n⚠️  Are you sure you want to rollback? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("❌ Rollback cancelled")
		return nil
	}

	// Perform rollback
	return runRollback(selectedVersion.Version)
}