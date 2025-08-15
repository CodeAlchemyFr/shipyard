package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/database"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Manage the SQLite database used for deployment versioning.`,
}

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database status and statistics",
	Long:  `Display information about the SQLite database including size, tables, and record counts.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDBStatus(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

var dbCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up old deployment records",
	Long:  `Remove old deployment records while keeping the most recent ones for each app.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDBCleanup(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbCleanupCmd)
}

func runDBStatus() error {
	db, err := database.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Get database file info
	info, err := os.Stat(db.GetPath())
	if err != nil {
		return fmt.Errorf("failed to get database file info: %w", err)
	}

	fmt.Printf("📊 Database Status\n\n")
	fmt.Printf("Database Path: %s\n", db.GetPath())
	fmt.Printf("Database Size: %.2f KB\n", float64(info.Size())/1024)
	fmt.Printf("Last Modified: %s\n\n", info.ModTime().Format("2006-01-02 15:04:05"))

	// Get table statistics
	conn := db.GetConnection()

	// Count apps
	var appCount int
	err = conn.QueryRow("SELECT COUNT(*) FROM apps").Scan(&appCount)
	if err != nil {
		return fmt.Errorf("failed to count apps: %w", err)
	}

	// Count deployments
	var deploymentCount int
	err = conn.QueryRow("SELECT COUNT(*) FROM deployments").Scan(&deploymentCount)
	if err != nil {
		return fmt.Errorf("failed to count deployments: %w", err)
	}

	// Count by status
	var successCount, failedCount, pendingCount int
	conn.QueryRow("SELECT COUNT(*) FROM deployments WHERE status = 'success'").Scan(&successCount)
	conn.QueryRow("SELECT COUNT(*) FROM deployments WHERE status = 'failed'").Scan(&failedCount)
	conn.QueryRow("SELECT COUNT(*) FROM deployments WHERE status = 'pending'").Scan(&pendingCount)

	fmt.Printf("📈 Statistics:\n")
	fmt.Printf("   Applications: %d\n", appCount)
	fmt.Printf("   Total Deployments: %d\n", deploymentCount)
	fmt.Printf("   ✅ Successful: %d\n", successCount)
	fmt.Printf("   ❌ Failed: %d\n", failedCount)
	fmt.Printf("   ⏳ Pending: %d\n\n", pendingCount)

	// Show recent apps with deployment counts
	if appCount > 0 {
		fmt.Printf("📱 Applications:\n")
		rows, err := conn.Query(`
			SELECT 
				a.name,
				COUNT(d.id) as deployment_count,
				MAX(d.deployed_at) as last_deployment
			FROM apps a
			LEFT JOIN deployments d ON a.id = d.app_id
			GROUP BY a.id, a.name
			ORDER BY last_deployment DESC
		`)
		if err != nil {
			return fmt.Errorf("failed to query app statistics: %w", err)
		}
		defer rows.Close()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintf(w, "   Name\tDeployments\tLast Deploy\n")
		fmt.Fprintf(w, "   ----\t-----------\t-----------\n")

		for rows.Next() {
			var name string
			var deployCount int
			var lastDeploy *string

			err := rows.Scan(&name, &deployCount, &lastDeploy)
			if err != nil {
				return fmt.Errorf("failed to scan app row: %w", err)
			}

			lastDeployStr := "Never"
			if lastDeploy != nil {
				lastDeployStr = *lastDeploy
				if len(lastDeployStr) > 16 {
					lastDeployStr = lastDeployStr[:16]
				}
			}

			fmt.Fprintf(w, "   %s\t%d\t%s\n", name, deployCount, lastDeployStr)
		}
		w.Flush()
	}

	return nil
}

func runDBCleanup() error {
	db, err := database.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	conn := db.GetConnection()

	fmt.Println("🧹 Cleaning up old deployment records...")

	// Keep only the last 20 deployments per app
	result, err := conn.Exec(`
		DELETE FROM deployments 
		WHERE id NOT IN (
			SELECT id FROM (
				SELECT id 
				FROM deployments d
				WHERE d.app_id = deployments.app_id
				ORDER BY deployed_at DESC
				LIMIT 20
			)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup deployments: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	fmt.Printf("✅ Cleanup complete. Removed %d old deployment records.\n", rowsAffected)
	fmt.Println("   Kept the most recent 20 deployments per application.")

	return nil
}