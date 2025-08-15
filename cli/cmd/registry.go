package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/registry"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage container registry credentials",
	Long:  `Add, list, remove, and manage container registry credentials for private images.`,
}

var registryAddCmd = &cobra.Command{
	Use:   "add [registry-url] [username] [password/token]",
	Short: "Add a new registry credential",
	Long: `Add credentials for a container registry.

Examples:
  shipyard registry add ghcr.io myuser ghp_token123
  shipyard registry add docker.io myuser mypassword
  shipyard registry add my-registry.com:5000 user token --email user@example.com
  shipyard registry add --default docker.io myuser mypass`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		registryURL := args[0]
		username := args[1]
		password := args[2]
		
		email, _ := cmd.Flags().GetString("email")
		registryType, _ := cmd.Flags().GetString("type")
		isDefault, _ := cmd.Flags().GetBool("default")

		if err := runRegistryAdd(registryURL, username, password, email, registryType, isDefault); err != nil {
			log.Fatalf("Failed to add registry: %v", err)
		}
	},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured registries",
	Long:  `Display all configured container registry credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRegistryList(); err != nil {
			log.Fatalf("Failed to list registries: %v", err)
		}
	},
}

var registryRemoveCmd = &cobra.Command{
	Use:   "remove [registry-url]",
	Short: "Remove a registry credential",
	Long:  `Remove credentials for a container registry.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registryURL := args[0]
		
		if err := runRegistryRemove(registryURL); err != nil {
			log.Fatalf("Failed to remove registry: %v", err)
		}
	},
}

var registryDefaultCmd = &cobra.Command{
	Use:   "default [registry-url]",
	Short: "Set a registry as default",
	Long:  `Set a registry as the default for image pulls.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registryURL := args[0]
		
		if err := runRegistryDefault(registryURL); err != nil {
			log.Fatalf("Failed to set default registry: %v", err)
		}
	},
}

func init() {
	// Add flags to registry add command
	registryAddCmd.Flags().String("email", "", "Email for registry authentication")
	registryAddCmd.Flags().String("type", "docker", "Registry type (docker, github, gitlab, aws)")
	registryAddCmd.Flags().Bool("default", false, "Set this registry as default")

	// Add subcommands
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryDefaultCmd)
}

func runRegistryAdd(registryURL, username, password, email, registryType string, isDefault bool) error {
	fmt.Printf("🔐 Adding registry credentials for %s...\n", registryURL)

	manager, err := registry.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	if err := manager.AddRegistry(registryURL, username, password, email, registryType, isDefault); err != nil {
		return fmt.Errorf("failed to add registry: %w", err)
	}

	fmt.Printf("✅ Registry %s added successfully", registryURL)
	if isDefault {
		fmt.Printf(" (set as default)")
	}
	fmt.Println()

	return nil
}

func runRegistryList() error {
	fmt.Println("📋 Container Registry Credentials:")

	manager, err := registry.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	registries, err := manager.ListRegistries()
	if err != nil {
		return fmt.Errorf("failed to list registries: %w", err)
	}

	if len(registries) == 0 {
		fmt.Println("   No registries configured.")
		fmt.Println("\n💡 Add a registry with: shipyard registry add <url> <username> <token>")
		return nil
	}

	fmt.Printf("%-30s %-15s %-10s %-10s %-20s\n", "Registry URL", "Username", "Type", "Default", "Created")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────")

	for _, reg := range registries {
		defaultMarker := ""
		if reg.IsDefault {
			defaultMarker = "✓"
		}

		fmt.Printf("%-30s %-15s %-10s %-10s %-20s\n",
			reg.RegistryURL,
			reg.Username,
			reg.RegistryType,
			defaultMarker,
			reg.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

func runRegistryRemove(registryURL string) error {
	fmt.Printf("🗑️  Removing registry %s...\n", registryURL)

	manager, err := registry.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	if err := manager.RemoveRegistry(registryURL); err != nil {
		return fmt.Errorf("failed to remove registry: %w", err)
	}

	fmt.Printf("✅ Registry %s removed successfully\n", registryURL)
	return nil
}

func runRegistryDefault(registryURL string) error {
	fmt.Printf("⭐ Setting %s as default registry...\n", registryURL)

	manager, err := registry.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	if err := manager.SetDefaultRegistry(registryURL); err != nil {
		return fmt.Errorf("failed to set default registry: %w", err)
	}

	fmt.Printf("✅ Registry %s set as default\n", registryURL)
	return nil
}