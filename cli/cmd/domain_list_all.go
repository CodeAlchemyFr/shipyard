package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/domains"
)

var domainListAllCmd = &cobra.Command{
	Use:   "list-all",
	Short: "List all domains across all applications",
	Long:  `Show all domains configured for all applications, grouped by base domain.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDomainListAll(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	domainCmd.AddCommand(domainListAllCmd)
}

func runDomainListAll() error {
	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Get all domains grouped by base domain
	domainGroups, err := domainManager.GetAllDomains()
	if err != nil {
		return fmt.Errorf("failed to get all domains: %w", err)
	}

	if len(domainGroups) == 0 {
		fmt.Printf("📋 No domains configured across all applications\n")
		fmt.Printf("💡 Add a domain with: shipyard domain add <hostname>\n")
		return nil
	}

	fmt.Printf("📋 All Domains Overview:\n\n")

	for _, group := range domainGroups {
		fmt.Printf("🌐 %s (Ingress: manifests/shared/%s.yaml)\n", group.BaseDomain, group.BaseDomain)
		
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintf(w, "   Hostname\tApp\tSSL\tCreated\n")
		fmt.Fprintf(w, "   --------\t---\t---\t-------\n")

		for _, domain := range group.Domains {
			sslStatus := "✅"
			if !domain.SSLEnabled {
				sslStatus = "❌"
			}
			createdAt := domain.CreatedAt.Format("2006-01-02")
			
			fmt.Fprintf(w, "   %s\t%s\t%s\t%s\n", 
				domain.Hostname, domain.AppName, sslStatus, createdAt)
		}
		
		w.Flush()
		fmt.Printf("   └─ SSL Certificate: %s-tls (wildcard)\n\n", group.BaseDomain)
	}

	// Summary statistics
	totalDomains := 0
	totalApps := make(map[string]bool)
	for _, group := range domainGroups {
		totalDomains += len(group.Domains)
		for _, domain := range group.Domains {
			totalApps[domain.AppName] = true
		}
	}

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   🌐 Base Domains: %d\n", len(domainGroups))
	fmt.Printf("   🔗 Total Hostnames: %d\n", totalDomains)
	fmt.Printf("   📱 Applications: %d\n", len(totalApps))

	return nil
}