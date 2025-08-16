package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/manifests"
)

var (
	deleteAll     bool
	forceDelete   bool
	confirmDelete bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete [app-name]",
	Short: "Delete an application and all its resources",
	Long: `Delete an application and clean up all associated resources:
- Kubernetes resources (deployment, service, ingress, secrets, etc.)
- Local manifest files
- Database entries

Examples:
  shipyard delete                    # Delete current app (from paas.yaml)
  shipyard delete hello-world       # Delete specific app
  shipyard delete --all             # Delete all apps
  shipyard delete --force           # Skip confirmation prompts`,
	Args: cobra.MaximumNArgs(1),
	Run:  func(cmd *cobra.Command, args []string) {
		runDelete(args)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all applications")
	deleteCmd.Flags().BoolVar(&forceDelete, "force", false, "Force deletion without confirmation")
	deleteCmd.Flags().BoolVar(&confirmDelete, "yes", false, "Automatically confirm deletion")
}

func runDelete(args []string) {
	var appName string
	var err error

	if deleteAll {
		err = deleteAllApps()
	} else if len(args) > 0 && args[0] != "" {
		// App name provided as argument
		appName = args[0]
		err = deleteApp(appName)
	} else {
		// Get app name from paas.yaml
		config, err := manifests.LoadConfig("paas.yaml")
		if err != nil {
			fmt.Printf("❌ Failed to load paas.yaml: %v\n", err)
			os.Exit(1)
		}
		appName = config.App.Name
		err = deleteApp(appName)
	}

	if err != nil {
		fmt.Printf("❌ Deletion failed: %v\n", err)
		os.Exit(1)
	}
}

func deleteApp(appName string) error {
	fmt.Printf("🗑️  Preparing to delete app: %s\n", appName)

	// Confirm deletion unless forced
	if !forceDelete && !confirmDelete {
		if !confirmDeletion(appName) {
			fmt.Println("❌ Deletion cancelled")
			return nil
		}
	}

	// Initialize version manager to get database connection
	vm := manifests.NewVersionManager(appName)
	defer vm.Close()

	// Delete Kubernetes resources
	fmt.Printf("☸️  Deleting Kubernetes resources for %s...\n", appName)
	if err := deleteKubernetesResources(appName); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete some Kubernetes resources: %v\n", err)
	}

	// Delete local manifest files
	fmt.Printf("📁 Cleaning up local manifest files...\n")
	if err := deleteManifestFiles(appName); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete manifest files: %v\n", err)
	}

	// Clean database entries
	fmt.Printf("🗃️  Cleaning up database entries...\n")
	if err := vm.DeleteApp(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to clean database: %v\n", err)
	}

	fmt.Printf("✅ Successfully deleted app: %s\n", appName)
	return nil
}

func deleteAllApps() error {
	fmt.Println("🗑️  Preparing to delete ALL applications")

	if !forceDelete && !confirmDelete {
		fmt.Print("⚠️  This will delete ALL applications and their data. Are you sure? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response != "y" && response != "yes" {
			fmt.Println("❌ Deletion cancelled")
			return nil
		}
	}

	// Get all apps from manifests directory
	manifestsDir := "manifests/apps"
	if _, err := os.Stat(manifestsDir); os.IsNotExist(err) {
		fmt.Println("ℹ️  No apps found to delete")
		return nil
	}

	entries, err := os.ReadDir(manifestsDir)
	if err != nil {
		return fmt.Errorf("failed to read manifests directory: %w", err)
	}

	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			appName := entry.Name()
			fmt.Printf("🗑️  Deleting app: %s\n", appName)
			if err := deleteApp(appName); err != nil {
				fmt.Printf("⚠️  Failed to delete %s: %v\n", appName, err)
			} else {
				deletedCount++
			}
		}
	}

	fmt.Printf("✅ Successfully deleted %d applications\n", deletedCount)
	return nil
}

func confirmDeletion(appName string) bool {
	fmt.Printf("⚠️  This will permanently delete the application '%s' and all its resources:\n", appName)
	fmt.Println("   - Kubernetes deployment, service, ingress, secrets")
	fmt.Println("   - Local manifest files")
	fmt.Println("   - Database entries and deployment history")
	fmt.Print("\nAre you sure you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	return response == "y" || response == "yes"
}

func deleteKubernetesResources(appName string) error {
	// Create a Kubernetes client
	client, err := manifests.CreateK8sClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Delete resources by applying manifests with --delete flag
	manifestDir := fmt.Sprintf("manifests/apps/%s", appName)
	if _, err := os.Stat(manifestDir); os.IsNotExist(err) {
		fmt.Printf("ℹ️  No manifest directory found for %s\n", appName)
		return nil
	}

	// Use kubectl delete command for simplicity
	return manifests.DeleteManifestsFromDirectory(client, manifestDir)
}

func deleteManifestFiles(appName string) error {
	manifestDir := fmt.Sprintf("manifests/apps/%s", appName)
	
	if _, err := os.Stat(manifestDir); os.IsNotExist(err) {
		fmt.Printf("ℹ️  No manifest files found for %s\n", appName)
		return nil
	}

	return os.RemoveAll(manifestDir)
}