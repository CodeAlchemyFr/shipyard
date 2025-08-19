package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/registry"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage container registry credentials",
	Long:  `Add, list, remove, and manage container registry credentials for private images.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Interactive mode when no subcommand is provided
		if err := runRegistryInteractive(); err != nil {
			log.Fatalf("Registry operation failed: %v", err)
		}
	},
}

var registryAddCmd = &cobra.Command{
	Use:   "add [registry-url] [username] [password/token]",
	Short: "Add a new registry credential",
	Long: `Add credentials for a container registry.

Examples:
  shipyard registry add ghcr.io myuser ghp_token123
  shipyard registry add docker.io myuser mypassword
  shipyard registry add --default docker.io myuser mypass`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		registryURL := args[0]
		username := args[1]
		password := args[2]
		
		isDefault, _ := cmd.Flags().GetBool("default")

		if err := runRegistryAdd(registryURL, username, password, isDefault); err != nil {
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
	registryAddCmd.Flags().Bool("default", false, "Set this registry as default")

	// Add subcommands
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryDefaultCmd)
}

func runRegistryAdd(registryURL, username, password string, isDefault bool) error {
	fmt.Printf("🔐 Adding registry credentials for %s...\n", registryURL)

	manager, err := registry.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	if err := manager.AddRegistry(registryURL, username, password, isDefault); err != nil {
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

	fmt.Printf("%-30s %-15s %-10s %-20s\n", "Registry URL", "Username", "Default", "Created")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")

	for _, reg := range registries {
		defaultMarker := ""
		if reg.IsDefault {
			defaultMarker = "✓"
		}

		fmt.Printf("%-30s %-15s %-10s %-20s\n",
			reg.RegistryURL,
			reg.Username,
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

// runRegistryInteractive provides an interactive menu for registry management
func runRegistryInteractive() error {
	for {
		fmt.Println("\n🐳 Registry Management")
		fmt.Println("===================")
		
		// Show current registries first
		manager, err := registry.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize registry manager: %w", err)
		}
		
		registries, err := manager.ListRegistries()
		if err != nil {
			manager.Close()
			return fmt.Errorf("failed to list registries: %w", err)
		}
		
		if len(registries) > 0 {
			fmt.Println("\nCurrent registries:")
			for i, reg := range registries {
				defaultMarker := ""
				if reg.IsDefault {
					defaultMarker = " (default)"
				}
				fmt.Printf("  %d. %s [%s]%s\n", i+1, reg.RegistryURL, reg.Username, defaultMarker)
			}
		} else {
			fmt.Println("\n📋 No registries configured")
		}
		
		manager.Close()
		
		fmt.Println("\nActions:")
		fmt.Println("  1. Add registry")
		fmt.Println("  2. Remove registry")
		fmt.Println("  3. Set default registry")
		fmt.Println("  4. List registries (detailed)")
		fmt.Println("  0. Exit")
		
		fmt.Print("\nSelect action: ")
		var choice string
		fmt.Scanln(&choice)
		
		switch strings.TrimSpace(choice) {
		case "1":
			if err := interactiveAddRegistry(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "2":
			if err := interactiveRemoveRegistry(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "3":
			if err := interactiveSetDefault(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "4":
			if err := runRegistryList(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "0", "":
			fmt.Println("👋 Goodbye!")
			return nil
		default:
			fmt.Println("❌ Invalid choice. Please select 0-4.")
		}
	}
}

// interactiveAddRegistry prompts user to add a registry interactively
func interactiveAddRegistry() error {
	fmt.Println("\n📝 Add Registry")
	fmt.Println("===============")
	
	var registryURL, username, password string
	var setDefault string
	
	fmt.Print("Registry URL (e.g., ghcr.io, docker.io): ")
	fmt.Scanln(&registryURL)
	
	if strings.TrimSpace(registryURL) == "" {
		return fmt.Errorf("registry URL cannot be empty")
	}
	
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username cannot be empty")
	}
	
	fmt.Print("Password/Token: ")
	fmt.Scanln(&password)
	
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password/token cannot be empty")
	}
	
	fmt.Print("Set as default? (y/N): ")
	fmt.Scanln(&setDefault)
	
	isDefault := strings.ToLower(strings.TrimSpace(setDefault)) == "y"
	
	return runRegistryAdd(registryURL, username, password, isDefault)
}

// interactiveRemoveRegistry prompts user to remove a registry
func interactiveRemoveRegistry() error {
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
		fmt.Println("📋 No registries to remove")
		return nil
	}
	
	fmt.Println("\n🗑️  Remove Registry")
	fmt.Println("==================")
	
	fmt.Println("Select registry to remove:")
	for i, reg := range registries {
		fmt.Printf("  %d. %s [%s]\n", i+1, reg.RegistryURL, reg.Username)
	}
	fmt.Println("  0. Cancel")
	
	fmt.Print("\nSelect: ")
	var choice string
	fmt.Scanln(&choice)
	
	if choice == "0" || strings.TrimSpace(choice) == "" {
		fmt.Println("❌ Cancelled")
		return nil
	}
	
	index, err := strconv.Atoi(strings.TrimSpace(choice))
	if err != nil || index < 1 || index > len(registries) {
		return fmt.Errorf("invalid selection")
	}
	
	selectedRegistry := registries[index-1]
	
	// Confirm removal
	fmt.Printf("⚠️  Are you sure you want to remove %s? (y/N): ", selectedRegistry.RegistryURL)
	var confirm string
	fmt.Scanln(&confirm)
	
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("❌ Cancelled")
		return nil
	}
	
	return runRegistryRemove(selectedRegistry.RegistryURL)
}

// interactiveSetDefault prompts user to set default registry
func interactiveSetDefault() error {
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
		fmt.Println("📋 No registries available")
		return nil
	}
	
	fmt.Println("\n⭐ Set Default Registry")
	fmt.Println("======================")
	
	fmt.Println("Select default registry:")
	for i, reg := range registries {
		defaultMarker := ""
		if reg.IsDefault {
			defaultMarker = " (current default)"
		}
		fmt.Printf("  %d. %s [%s]%s\n", i+1, reg.RegistryURL, reg.Username, defaultMarker)
	}
	fmt.Println("  0. Cancel")
	
	fmt.Print("\nSelect: ")
	var choice string
	fmt.Scanln(&choice)
	
	if choice == "0" || strings.TrimSpace(choice) == "" {
		fmt.Println("❌ Cancelled")
		return nil
	}
	
	index, err := strconv.Atoi(strings.TrimSpace(choice))
	if err != nil || index < 1 || index > len(registries) {
		return fmt.Errorf("invalid selection")
	}
	
	selectedRegistry := registries[index-1]
	
	if selectedRegistry.IsDefault {
		fmt.Printf("✅ %s is already the default registry\n", selectedRegistry.RegistryURL)
		return nil
	}
	
	return runRegistryDefault(selectedRegistry.RegistryURL)
}