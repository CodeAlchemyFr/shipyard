package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/manifests"
)

var (
	releasesLimit int
)

var releasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "Show deployment release history",
	Long:  `Display the history of deployments with versions, images, and status.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runReleases(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	releasesCmd.Flags().IntVar(&releasesLimit, "limit", 10, "Number of releases to show")
}

func runReleases() error {
	// Parse current config to get app name
	config, err := manifests.LoadConfig("paas.yaml")
	if err != nil {
		return fmt.Errorf("failed to load paas.yaml: %w", err)
	}

	// Load version manager
	versionManager := manifests.NewVersionManager(config.App.Name)

	// Get version history
	versions, err := versionManager.ListVersions(releasesLimit)
	if err != nil {
		return fmt.Errorf("failed to load version history: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("📋 No deployment history found for app: %s\n", config.App.Name)
		return nil
	}

	fmt.Printf("📋 Deployment History for %s:\n\n", config.App.Name)

	// Table header
	fmt.Printf("┌%-12s┬%-20s┬%-15s┬%-10s┬%-20s┬%-15s┐\n", 
		strings.Repeat("─", 12), strings.Repeat("─", 20), strings.Repeat("─", 15), 
		strings.Repeat("─", 10), strings.Repeat("─", 20), strings.Repeat("─", 15))
	fmt.Printf("│%-12s│%-20s│%-15s│%-10s│%-20s│%-15s│\n", 
		"VERSION", "IMAGE TAG", "STATUS", "AGE", "DEPLOYED AT", "ROLLBACK FROM")
	fmt.Printf("├%-12s┼%-20s┼%-15s┼%-10s┼%-20s┼%-15s┤\n", 
		strings.Repeat("─", 12), strings.Repeat("─", 20), strings.Repeat("─", 15), 
		strings.Repeat("─", 10), strings.Repeat("─", 20), strings.Repeat("─", 15))

	// Current deployment indicator
	for i, version := range versions {
		statusIcon := getStatusIcon(version.Status)
		age := formatAge(time.Since(version.Timestamp))
		deployedAt := version.Timestamp.Format("2006-01-02 15:04")
		
		// Truncate long image tags
		imageTag := version.ImageTag
		if len(imageTag) > 18 {
			imageTag = imageTag[:15] + "..."
		}
		
		// Current deployment marker
		currentMarker := ""
		if i == 0 && version.Status == "success" {
			currentMarker = " (current)"
		}
		
		rollbackFrom := ""
		if version.RollbackTo != "" {
			rollbackFrom = version.RollbackTo
			if len(rollbackFrom) > 13 {
				rollbackFrom = rollbackFrom[:10] + "..."
			}
		}

		fmt.Printf("│%-12s│%-20s│%-15s│%-10s│%-20s│%-15s│\n", 
			version.Version + currentMarker, 
			imageTag, 
			statusIcon + " " + version.Status, 
			age, 
			deployedAt,
			rollbackFrom)
	}

	fmt.Printf("└%-12s┴%-20s┴%-15s┴%-10s┴%-20s┴%-15s┘\n", 
		strings.Repeat("─", 12), strings.Repeat("─", 20), strings.Repeat("─", 15), 
		strings.Repeat("─", 10), strings.Repeat("─", 20), strings.Repeat("─", 15))

	fmt.Printf("\n💡 Usage:\n")
	fmt.Printf("   shipyard rollback %s    # Rollback to specific version\n", versions[1].Version)
	fmt.Printf("   shipyard rollback %s        # Rollback to specific image tag\n", versions[1].ImageTag)
	fmt.Printf("   shipyard rollback                 # Rollback to latest successful\n")

	return nil
}

func getStatusIcon(status string) string {
	switch status {
	case "success":
		return "✅"
	case "failed":
		return "❌"
	case "pending":
		return "⏳"
	default:
		return "❓"
	}
}

func formatAge(duration time.Duration) string {
	if duration < time.Minute {
		return "< 1m"
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}