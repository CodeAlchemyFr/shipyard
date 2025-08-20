package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/manifests"
	"github.com/shipyard/cli/pkg/k8s"
	versionpkg "github.com/shipyard/cli/pkg/version"
)

var (
	autoRegistry bool // Flag for auto-selecting registry secrets
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application to Kubernetes",
	Long: `Deploy your application by generating Kubernetes manifests and applying them to the cluster.
This will create:
- A deployment.yaml for your application
- A secrets.yaml for environment variables (base64 encoded)  
- A service.yaml for internal load balancing
- Update shared ingress files for domains

By default, you'll be prompted to select which registry secrets to use.
Use --auto-registry to automatically select the best matching registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDeploy(); err != nil {
			log.Fatalf("Deploy failed: %v", err)
		}
	},
}

func runDeploy() error {
	// Check for updates (non-blocking)
	go versionpkg.NotifyIfUpdateAvailable(versionpkg.Current)
	
	fmt.Println("🚀 Starting deployment...")

	// 1. Parse paas.yaml configuration
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// 1.5. Validate DNS names and ask for confirmation if needed
	if err := validateAndConfirmDNSNames(config); err != nil {
		return fmt.Errorf("DNS validation failed: %w", err)
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
	generator := manifests.NewGeneratorWithVersionAndMode(config, deployVersion, !autoRegistry)
	
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

	fmt.Printf("🔧 Applying manifests for %s...\n", config.App.Name)
	if err := client.ApplyManifests(config.App.Name); err != nil {
		// Mark deployment as failed
		versionManager.UpdateVersionStatus(deployVersion.Version, "failed")
		
		// Try to get some diagnostic information
		fmt.Println("\n🔍 Diagnostic information:")
		if pods, podErr := client.GetPods(config.App.Name); podErr == nil {
			for _, pod := range pods {
				fmt.Printf("   Pod %s: %s\n", pod.Name, pod.Status.Phase)
				if pod.Status.Phase == "Failed" || pod.Status.Phase == "Pending" {
					for _, condition := range pod.Status.Conditions {
						if condition.Status == "False" {
							fmt.Printf("     - %s: %s\n", condition.Type, condition.Message)
						}
					}
				}
			}
		}
		
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	// Mark deployment as successful
	if err := versionManager.UpdateVersionStatus(deployVersion.Version, "success"); err != nil {
		fmt.Printf("⚠️  Warning: failed to update version status: %v\n", err)
	}

	// Update deployment for CI/CD if enabled (after successful deployment)
	if config.CICD.Enabled {
		fmt.Printf("🔄 Updating deployment for CI/CD mode...\n")
		if err := generator.UpdateDeploymentForCICD(); err != nil {
			fmt.Printf("⚠️  Warning: failed to update deployment for CI/CD: %v\n", err)
		} else {
			fmt.Printf("✅ Deployment updated for CI/CD - image replaced with ${IMAGE_TAG}\n")
		}
	}

	fmt.Printf("✅ Deployment successful!\n")
	fmt.Printf("   App: %s\n", config.App.Name)
	fmt.Printf("   Version: %s\n", deployVersion.Version)
	fmt.Printf("   Image: %s\n", config.App.Image)
	
	// Offer to show logs
	fmt.Printf("\n💡 To follow logs, run: shipyard logs %s -f\n", config.App.Name)
	fmt.Printf("💡 To check status, run: shipyard status\n")
	
	return nil
}

// validateAndConfirmDNSNames checks if names need DNS normalization and asks for user confirmation
func validateAndConfirmDNSNames(config *manifests.Config) error {
	// Check if app name is DNS compliant
	originalName := config.App.Name
	normalizedName := normalizeDNSName(originalName)
	
	changes := []string{}
	if originalName != normalizedName {
		changes = append(changes, fmt.Sprintf("App name: '%s' → '%s'", originalName, normalizedName))
	}
	
	// Check namespace if specified
	if config.App.Namespace != "" {
		originalNamespace := config.App.Namespace
		normalizedNamespace := normalizeDNSName(originalNamespace)
		if originalNamespace != normalizedNamespace {
			changes = append(changes, fmt.Sprintf("Namespace: '%s' → '%s'", originalNamespace, normalizedNamespace))
		}
	}
	
	// If no changes needed, continue
	if len(changes) == 0 {
		return nil
	}
	
	// Show proposed changes and ask for confirmation
	fmt.Println("⚠️  Les noms suivants ne sont pas compatibles DNS-1035 et seront modifiés:")
	for _, change := range changes {
		fmt.Printf("   %s\n", change)
	}
	fmt.Println("\n📋 Règles DNS-1035:")
	fmt.Println("   - Uniquement lettres minuscules, chiffres et tirets (-)")
	fmt.Println("   - Doit commencer par une lettre")
	fmt.Println("   - Doit finir par une lettre ou un chiffre")
	
	fmt.Print("\n❓ Continuer avec ces modifications ? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" && response != "oui" {
		fmt.Println("❌ Déploiement annulé")
		fmt.Println("💡 Modifiez votre paas.yaml pour utiliser des noms compatibles DNS")
		return fmt.Errorf("user cancelled deployment due to DNS name changes")
	}
	
	fmt.Println("✅ Modifications acceptées, déploiement en cours...")
	return nil
}

// normalizeDNSName converts a string to be DNS-1035 compliant
func normalizeDNSName(name string) string {
	// Convert to lowercase and replace underscores with hyphens
	result := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	
	// Remove any characters that aren't alphanumeric or hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	result = reg.ReplaceAllString(result, "")
	
	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "app-" + result
	}
	
	// Ensure it doesn't start or end with hyphen
	result = strings.Trim(result, "-")
	
	// If empty after cleaning, use a default
	if result == "" {
		result = "my-app"
	}
	
	return result
}

func init() {
	deployCmd.Flags().BoolVar(&autoRegistry, "auto-registry", false, "Automatically select the best matching registry instead of prompting")
}