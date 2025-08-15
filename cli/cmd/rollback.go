package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/manifests"
	"github.com/shipyard/cli/pkg/k8s"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [version|image-tag]",
	Short: "Rollback to a previous deployment version",
	Long: `Rollback your application to a previous deployment version.
You can specify either a version (e.g., v1634567890) or an image tag (e.g., v1.2.3).
If no version is specified, it will rollback to the latest successful deployment.`,
	Run: func(cmd *cobra.Command, args []string) {
		var targetVersion string
		if len(args) > 0 {
			targetVersion = args[0]
		}
		
		if err := runRollback(targetVersion); err != nil {
			log.Fatalf("Rollback failed: %v", err)
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