package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/domains"
	"github.com/shipyard/cli/pkg/manifests"
)

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage domains for applications",
	Long:  `Add, remove, and list domains for your applications.`,
}

var domainAddCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a domain to the current application",
	Long: `Add a domain to the current application's paas.yaml configuration.
This will update the paas.yaml file and regenerate the ingress.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		if err := runDomainAdd(domain); err != nil {
			log.Fatalf("Failed to add domain: %v", err)
		}
	},
}

var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains for the current application",
	Long:  `Show all domains configured for the current application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDomainList(); err != nil {
			log.Fatalf("Failed to list domains: %v", err)
		}
	},
}

var domainRemoveCmd = &cobra.Command{
	Use:   "remove <domain>",
	Short: "Remove a domain from the current application",
	Long: `Remove a domain from the current application's paas.yaml configuration.
This will update the paas.yaml file and regenerate the ingress.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		if err := runDomainRemove(domain); err != nil {
			log.Fatalf("Failed to remove domain: %v", err)
		}
	},
}

func init() {
	domainCmd.AddCommand(domainAddCmd)
	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainRemoveCmd)
}

func runDomainAdd(hostname string) error {
	// Load current config to get app name
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Add domain to database
	if err := domainManager.AddDomain(config.App.Name, hostname); err != nil {
		return fmt.Errorf("failed to add domain: %w", err)
	}

	fmt.Printf("✅ Added domain: %s → %s\n", hostname, config.App.Name)
	fmt.Printf("💾 Saved to database\n")

	// Regenerate all ingress files
	fmt.Println("🌐 Regenerating ingress configuration...")
	generator := manifests.NewGenerator(config)
	if err := generator.GenerateIngressFromDatabase(); err != nil {
		return fmt.Errorf("failed to regenerate ingress: %w", err)
	}

	fmt.Printf("🚀 To apply changes to cluster, run: shipyard deploy\n")
	return nil
}

func runDomainList() error {
	// Load current config to get app name
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Get domains for current app
	appDomains, err := domainManager.GetDomainsForApp(config.App.Name)
	if err != nil {
		return fmt.Errorf("failed to get domains: %w", err)
	}

	if len(appDomains) == 0 {
		fmt.Printf("📋 No domains configured for app: %s\n", config.App.Name)
		fmt.Printf("💡 Add a domain with: shipyard domain add <hostname>\n")
		return nil
	}

	fmt.Printf("📋 Domains for app %s:\n\n", config.App.Name)

	// Group domains by base domain
	domainGroups := make(map[string][]domains.Domain)
	for _, domain := range appDomains {
		domainGroups[domain.BaseDomain] = append(domainGroups[domain.BaseDomain], domain)
	}

	for baseDomain, domainList := range domainGroups {
		fmt.Printf("🌐 %s (Ingress: manifests/shared/%s.yaml)\n", baseDomain, baseDomain)
		for _, domain := range domainList {
			sslStatus := "✅"
			if !domain.SSLEnabled {
				sslStatus = "❌"
			}
			fmt.Printf("   ├─ https://%s %s\n", domain.Hostname, sslStatus)
		}
		fmt.Printf("   └─ SSL: %s-tls (wildcard)\n\n", baseDomain)
	}

	return nil
}

func runDomainRemove(hostname string) error {
	// Load current config to get app name
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Remove domain from database
	if err := domainManager.RemoveDomain(config.App.Name, hostname); err != nil {
		return fmt.Errorf("failed to remove domain: %w", err)
	}

	fmt.Printf("✅ Removed domain: %s from %s\n", hostname, config.App.Name)
	fmt.Printf("💾 Updated database\n")

	// Regenerate all ingress files
	fmt.Println("🌐 Regenerating ingress configuration...")
	generator := manifests.NewGenerator(config)
	if err := generator.GenerateIngressFromDatabase(); err != nil {
		return fmt.Errorf("failed to regenerate ingress: %w", err)
	}

	// Cleanup orphaned ingress files
	if err := generator.CleanupIngressFiles(); err != nil {
		fmt.Printf("⚠️  Warning: failed to cleanup ingress files: %v\n", err)
	}

	fmt.Printf("🚀 To apply changes to cluster, run: shipyard deploy\n")
	return nil
}

