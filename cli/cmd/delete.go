package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/config"
	"github.com/shipyard/cli/pkg/domains"
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

	// Delete Kubernetes resources (including ingress)
	fmt.Printf("☸️  Deleting Kubernetes resources for %s...\n", appName)
	if err := deleteKubernetesResources(appName); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete some Kubernetes resources: %v\n", err)
	}

	// Delete shared ingress files for this app's domains
	fmt.Printf("🌐 Cleaning up ingress files...\n")
	if err := deleteIngressFiles(appName); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete ingress files: %v\n", err)
	}

	// Delete local manifest files (just the app directory)
	fmt.Printf("📁 Cleaning up app manifest files...\n")
	if err := deleteManifestFiles(appName); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete manifest files: %v\n", err)
	}

	// Clean database entries (this will also clean up domain associations)
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
	manifestsDir, err := config.GetAppsDir()
	if err != nil {
		return fmt.Errorf("failed to get apps directory: %w", err)
	}
	
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

	// Clean up empty directories after deleting all apps
	fmt.Printf("🧹 Cleaning up empty directories...\n")
	manifestsBaseDir, _ := config.GetManifestsDir()
	cleanupEmptyDirectories(manifestsBaseDir)
	
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

	// Get apps directory from global config
	appsDir, err := config.GetAppsDir()
	if err != nil {
		return fmt.Errorf("failed to get apps directory: %w", err)
	}

	// Delete resources by applying manifests with --delete flag
	manifestDir := filepath.Join(appsDir, appName)
	if _, err := os.Stat(manifestDir); os.IsNotExist(err) {
		fmt.Printf("ℹ️  No manifest directory found for %s\n", appName)
		return nil
	}

	// Use kubectl delete command for simplicity
	return manifests.DeleteManifestsFromDirectory(client, manifestDir)
}

func deleteManifestFiles(appName string) error {
	// Get apps directory from global config
	appsDir, err := config.GetAppsDir()
	if err != nil {
		return fmt.Errorf("failed to get apps directory: %w", err)
	}

	manifestDir := filepath.Join(appsDir, appName)
	
	// Check if directory exists
	if _, err := os.Stat(manifestDir); os.IsNotExist(err) {
		fmt.Printf("ℹ️  No manifest files found for %s\n", appName)
		return nil
	}

	// Show what we're about to delete
	fmt.Printf("🗑️  Deleting manifest directory: %s\n", manifestDir)
	
	// List files being deleted for debug
	files, err := os.ReadDir(manifestDir)
	if err == nil {
		for _, file := range files {
			fmt.Printf("   - %s\n", file.Name())
		}
	}

	// Remove the directory
	err = os.RemoveAll(manifestDir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", manifestDir, err)
	}

	fmt.Printf("✅ Deleted manifest directory: %s\n", manifestDir)
	
	// Clean up empty parent directories
	manifestsBaseDir, _ := config.GetManifestsDir()
	cleanupEmptyDirectories(manifestsBaseDir)
	
	return nil
}

// deleteIngressFiles removes ingress files for domains associated with the app
func deleteIngressFiles(appName string) error {
	// Get domains manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Get domains for this app
	appDomains, err := domainManager.GetDomainsForApp(appName)
	if err != nil {
		return fmt.Errorf("failed to get domains for app %s: %w", appName, err)
	}

	if len(appDomains) == 0 {
		fmt.Printf("ℹ️  No domains found for app %s\n", appName)
		return nil
	}

	// Get shared directory
	sharedDir, err := config.GetSharedDir()
	if err != nil {
		return fmt.Errorf("failed to get shared directory: %w", err)
	}

	// Group domains by base domain and check if we need to delete ingress files
	baseDomains := make(map[string]bool)
	for _, domain := range appDomains {
		baseDomains[domain.BaseDomain] = true
	}

	// For each base domain, check if there are other apps using it
	for baseDomain := range baseDomains {
		// Get all domains for this base domain
		allDomainsForBase, err := domainManager.GetDomainsByBaseDomain(baseDomain)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to get domains for base domain %s: %v\n", baseDomain, err)
			continue
		}

		// Check if all domains for this base domain belong to the app being deleted
		shouldDeleteIngress := true
		for _, domain := range allDomainsForBase {
			if domain.AppName != appName {
				shouldDeleteIngress = false
				break
			}
		}

		if shouldDeleteIngress {
			// Delete the ingress file
			ingressFile := filepath.Join(sharedDir, fmt.Sprintf("%s.yaml", baseDomain))
			if _, err := os.Stat(ingressFile); err == nil {
				if err := os.Remove(ingressFile); err != nil {
					fmt.Printf("⚠️  Warning: Failed to remove ingress file %s: %v\n", ingressFile, err)
				} else {
					fmt.Printf("🗑️  Removed ingress: %s\n", ingressFile)
				}
			}

			// Also delete from Kubernetes
			if err := deleteIngressFromKubernetes(baseDomain); err != nil {
				fmt.Printf("⚠️  Warning: Failed to delete ingress from Kubernetes: %v\n", err)
			}
		} else {
			fmt.Printf("ℹ️  Keeping ingress for %s (used by other apps)\n", baseDomain)
		}
	}

	return nil
}

// deleteIngressFromKubernetes removes ingress from Kubernetes cluster
func deleteIngressFromKubernetes(baseDomain string) error {
	ingressName := fmt.Sprintf("%s-ingress", baseDomain)
	
	fmt.Printf("🗑️  Deleting ingress %s from Kubernetes\n", ingressName)
	
	// Use kubectl to delete the ingress directly
	cmd := fmt.Sprintf("kubectl delete ingress %s --ignore-not-found=true", ingressName)
	
	// Execute the command (this is a simple approach)
	// In a production environment, you'd use proper subprocess execution
	if err := executeKubectlCommand(cmd); err != nil {
		fmt.Printf("⚠️  Warning: Could not delete ingress %s: %v\n", ingressName, err)
	}
	
	return nil
}

// executeKubectlCommand executes a kubectl command
func executeKubectlCommand(cmdStr string) error {
	fmt.Printf("📋 Executing: %s\n", cmdStr)
	
	// Split the command string into parts
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	
	// Execute the command
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("command failed: %s, output: %s", err, string(output))
	}
	
	if len(output) > 0 {
		fmt.Printf("✅ %s\n", strings.TrimSpace(string(output)))
	}
	
	return nil
}

// cleanupEmptyDirectories removes empty directories recursively starting from the given path
func cleanupEmptyDirectories(startPath string) {
	// Check if the directory exists
	if _, err := os.Stat(startPath); os.IsNotExist(err) {
		return
	}

	// Read the directory contents
	entries, err := os.ReadDir(startPath)
	if err != nil {
		return
	}

	// Recursively clean subdirectories first
	for _, entry := range entries {
		if entry.IsDir() {
			subPath := fmt.Sprintf("%s/%s", startPath, entry.Name())
			cleanupEmptyDirectories(subPath)
		}
	}

	// Re-read directory contents after cleaning subdirectories
	entries, err = os.ReadDir(startPath)
	if err != nil {
		return
	}

	// If directory is now empty, remove it (but not the root manifests directory)
	manifestsBaseDir, _ := config.GetManifestsDir()
	if len(entries) == 0 && startPath != manifestsBaseDir {
		fmt.Printf("🧹 Removing empty directory: %s\n", startPath)
		os.Remove(startPath)
	}
}