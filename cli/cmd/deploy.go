package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/manifests"
	"github.com/shipyard/cli/pkg/k8s"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application to Kubernetes",
	Long: `Deploy your application by generating Kubernetes manifests and applying them to the cluster.
This will create:
- A deployment.yaml for your application
- A secrets.yaml for environment variables (base64 encoded)  
- A service.yaml for internal load balancing
- Update shared ingress files for domains`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDeploy(); err != nil {
			log.Fatalf("Deploy failed: %v", err)
		}
	},
}

func runDeploy() error {
	fmt.Println("🚀 Starting deployment...")

	// 1. Parse paas.yaml configuration
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// 2. Create version manager and generate new version
	versionManager := manifests.NewVersionManager(config.App.Name)
	deployVersion, err := versionManager.GenerateVersion(config)
	if err != nil {
		return fmt.Errorf("failed to generate version: %w", err)
	}

	// Save the deployment version
	if err := versionManager.SaveVersion(deployVersion); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	// 3. Generate manifests for the application with version tracking
	generator := manifests.NewGeneratorWithVersion(config, deployVersion)
	
	fmt.Printf("📦 Generating manifests for app: %s (version: %s)\n", config.App.Name, deployVersion.Version)
	
	if err := generator.GenerateAppManifests(); err != nil {
		// Mark deployment as failed
		versionManager.UpdateVersionStatus(deployVersion.Version, "failed")
		return fmt.Errorf("failed to generate app manifests: %w", err)
	}

	// 4. Update shared ingress files
	fmt.Println("🌐 Updating ingress configuration...")
	if err := generator.UpdateIngressManifests(); err != nil {
		// Mark deployment as failed
		versionManager.UpdateVersionStatus(deployVersion.Version, "failed")
		return fmt.Errorf("failed to update ingress: %w", err)
	}

	// 5. Apply manifests to Kubernetes
	fmt.Println("☸️  Applying to Kubernetes cluster...")
	client, err := k8s.NewClient()
	if err != nil {
		// Mark deployment as failed
		versionManager.UpdateVersionStatus(deployVersion.Version, "failed")
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	if err := client.ApplyManifests(config.App.Name); err != nil {
		// Mark deployment as failed
		versionManager.UpdateVersionStatus(deployVersion.Version, "failed")
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	// Mark deployment as successful
	if err := versionManager.UpdateVersionStatus(deployVersion.Version, "success"); err != nil {
		fmt.Printf("⚠️  Warning: failed to update version status: %v\n", err)
	}

	fmt.Printf("✅ Deployment successful!\n")
	fmt.Printf("   App: %s\n", config.App.Name)
	fmt.Printf("   Version: %s\n", deployVersion.Version)
	fmt.Printf("   Image: %s\n", config.App.Image)
	return nil
}