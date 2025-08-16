package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/monitoring"
)

var healthCmd = &cobra.Command{
	Use:   "health [app-name]",
	Short: "Check health status of applications",
	Long: `Check the health status of your deployed applications.

This command shows:
- HTTP health check results
- Pod readiness status
- Service availability
- Recent health check history

Examples:
  shipyard health                     # Check health of all apps
  shipyard health my-app              # Check health of specific app
  shipyard health --watch            # Watch health status continuously
  shipyard health --history 1h       # Show health history for last hour`,
	Run: func(cmd *cobra.Command, args []string) {
		var appName string
		if len(args) > 0 {
			appName = args[0]
		}

		watch, _ := cmd.Flags().GetBool("watch")
		history, _ := cmd.Flags().GetDuration("history")

		if err := runHealth(appName, watch, history); err != nil {
			log.Fatalf("Health command failed: %v", err)
		}
	},
}

func init() {
	healthCmd.Flags().BoolP("watch", "w", false, "Watch health status continuously")
	healthCmd.Flags().DurationP("history", "t", 0, "Show health check history for specified duration")
}

func runHealth(appName string, watch bool, history time.Duration) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	if watch {
		return runHealthWatch(collector, appName)
	}

	if history > 0 {
		return showHealthHistory(collector, appName, history)
	}

	return showCurrentHealth(collector, appName)
}

func showCurrentHealth(collector *monitoring.Collector, appName string) error {
	healthChecks, err := getCurrentHealthChecks(collector, appName)
	if err != nil {
		return fmt.Errorf("failed to get health checks: %w", err)
	}

	displayHealthTable(healthChecks, appName)
	return nil
}

func runHealthWatch(collector *monitoring.Collector, appName string) error {
	fmt.Printf("🔍 Watching health status (press Ctrl+C to stop)\n")
	fmt.Println()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial check
	if err := showCurrentHealth(collector, appName); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			fmt.Print("\033[2J\033[H") // Clear screen
			fmt.Printf("🔍 Health Status - %s (auto-refresh: 10s)\n\n", time.Now().Format("15:04:05"))
			if err := showCurrentHealth(collector, appName); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

func showHealthHistory(collector *monitoring.Collector, appName string, duration time.Duration) error {
	history, err := getHealthHistory(collector, appName, duration)
	if err != nil {
		return fmt.Errorf("failed to get health history: %w", err)
	}

	displayHealthHistory(history, appName, duration)
	return nil
}

// HealthCheckResult represents a health check result
type HealthCheckResult struct {
	AppName      string
	Endpoint     string
	Status       string
	StatusCode   int
	ResponseTime int
	ErrorMessage string
	LastCheck    time.Time
	Uptime       float64
}

// HealthHistoryItem represents a historical health check
type HealthHistoryItem struct {
	AppName      string
	Endpoint     string
	Status       string
	StatusCode   int
	ResponseTime int
	CheckedAt    time.Time
}

func getCurrentHealthChecks(collector *monitoring.Collector, appName string) ([]HealthCheckResult, error) {
	// This would normally query the database for latest health checks
	// For now, return sample data
	
	results := []HealthCheckResult{
		{
			AppName:      "web-app",
			Endpoint:     "/health",
			Status:       "healthy",
			StatusCode:   200,
			ResponseTime: 45,
			LastCheck:    time.Now().Add(-30 * time.Second),
			Uptime:       99.8,
		},
		{
			AppName:      "api-service",
			Endpoint:     "/health",
			Status:       "unhealthy",
			StatusCode:   503,
			ResponseTime: 2000,
			ErrorMessage: "Service unavailable",
			LastCheck:    time.Now().Add(-15 * time.Second),
			Uptime:       95.2,
		},
		{
			AppName:      "worker",
			Endpoint:     "/ping",
			Status:       "healthy",
			StatusCode:   200,
			ResponseTime: 12,
			LastCheck:    time.Now().Add(-45 * time.Second),
			Uptime:       100.0,
		},
	}

	// Filter by app name if specified
	if appName != "" {
		var filtered []HealthCheckResult
		for _, result := range results {
			if result.AppName == appName {
				filtered = append(filtered, result)
			}
		}
		return filtered, nil
	}

	return results, nil
}

func getHealthHistory(collector *monitoring.Collector, appName string, duration time.Duration) ([]HealthHistoryItem, error) {
	// This would normally query the database for historical health checks
	// For now, return sample data
	
	now := time.Now()
	var history []HealthHistoryItem
	
	// Generate sample historical data
	for i := 0; i < 20; i++ {
		checkTime := now.Add(-time.Duration(i*5) * time.Minute)
		if checkTime.Before(now.Add(-duration)) {
			break
		}
		
		status := "healthy"
		statusCode := 200
		responseTime := 50 + i*2
		
		// Add some failures for demonstration
		if i%7 == 0 {
			status = "unhealthy"
			statusCode = 503
			responseTime = 2000
		}
		
		history = append(history, HealthHistoryItem{
			AppName:      "web-app",
			Endpoint:     "/health",
			Status:       status,
			StatusCode:   statusCode,
			ResponseTime: responseTime,
			CheckedAt:    checkTime,
		})
	}

	return history, nil
}

func displayHealthTable(results []HealthCheckResult, appName string) {
	if len(results) == 0 {
		fmt.Println("🏥 No health check data found")
		return
	}

	title := "🏥 Application Health Status"
	if appName != "" {
		title = fmt.Sprintf("🏥 Health Status for %s", appName)
	}
	
	fmt.Println(title)
	fmt.Println("=" + fmt.Sprintf("%*s", len(title)-1, ""))
	fmt.Println()

	// Header
	fmt.Printf("┌%-15s┬%-12s┬%-10s┬%-6s┬%-12s┬%-8s┬%-12s┐\n",
		"───────────────", "────────────", "──────────", "──────", "────────────", "────────", "────────────")
	fmt.Printf("│%-15s│%-12s│%-10s│%-6s│%-12s│%-8s│%-12s│\n",
		"APP NAME", "ENDPOINT", "STATUS", "CODE", "RESP TIME", "UPTIME", "LAST CHECK")
	fmt.Printf("├%-15s┼%-12s┼%-10s┼%-6s┼%-12s┼%-8s┼%-12s┤\n",
		"───────────────", "────────────", "──────────", "──────", "────────────", "────────", "────────────")

	// Data rows
	for _, result := range results {
		statusIcon := "🟢"
		statusText := result.Status
		if result.Status != "healthy" {
			statusIcon = "🔴"
		}

		responseTimeStr := fmt.Sprintf("%dms", result.ResponseTime)
		uptimeStr := fmt.Sprintf("%.1f%%", result.Uptime)
		lastCheckStr := formatTimeAgo(result.LastCheck)

		fmt.Printf("│%-15s│%-12s│%s %-8s│%-6d│%-12s│%-8s│%-12s│\n",
			truncateString(result.AppName, 15),
			truncateString(result.Endpoint, 12),
			statusIcon,
			truncateString(statusText, 7),
			result.StatusCode,
			responseTimeStr,
			uptimeStr,
			lastCheckStr,
		)

		// Show error message if available
		if result.ErrorMessage != "" {
			fmt.Printf("│%-15s│%-12s│%-10s│%-6s│%-12s│%-8s│%-12s│\n",
				"", "", fmt.Sprintf("└─ %s", truncateString(result.ErrorMessage, 45)), "", "", "", "")
		}
	}

	fmt.Printf("└%-15s┴%-12s┴%-10s┴%-6s┴%-12s┴%-8s┴%-12s┘\n",
		"───────────────", "────────────", "──────────", "──────", "────────────", "────────", "────────────")
	
	fmt.Printf("\n💡 Tip: Use --watch to monitor health continuously\n")
}

func displayHealthHistory(history []HealthHistoryItem, appName string, duration time.Duration) {
	if len(history) == 0 {
		fmt.Printf("🏥 No health history found for the last %v\n", duration)
		return
	}

	title := fmt.Sprintf("🏥 Health History (Last %v)", duration)
	if appName != "" {
		title = fmt.Sprintf("🏥 Health History for %s (Last %v)", appName, duration)
	}
	
	fmt.Println(title)
	fmt.Println("=" + fmt.Sprintf("%*s", len(title)-1, ""))
	fmt.Println()

	// Header
	fmt.Printf("┌%-20s┬%-10s┬%-6s┬%-12s┬%-12s┐\n",
		"────────────────────", "──────────", "──────", "────────────", "────────────")
	fmt.Printf("│%-20s│%-10s│%-6s│%-12s│%-12s│\n",
		"TIMESTAMP", "STATUS", "CODE", "RESP TIME", "ENDPOINT")
	fmt.Printf("├%-20s┼%-10s┼%-6s┼%-12s┼%-12s┤\n",
		"────────────────────", "──────────", "──────", "────────────", "────────────")

	// Data rows
	for _, item := range history {
		statusIcon := "🟢"
		if item.Status != "healthy" {
			statusIcon = "🔴"
		}

		timestampStr := item.CheckedAt.Format("15:04:05 Jan 02")
		responseTimeStr := fmt.Sprintf("%dms", item.ResponseTime)

		fmt.Printf("│%-20s│%s %-8s│%-6d│%-12s│%-12s│\n",
			timestampStr,
			statusIcon,
			truncateString(item.Status, 7),
			item.StatusCode,
			responseTimeStr,
			truncateString(item.Endpoint, 12),
		)
	}

	fmt.Printf("└%-20s┴%-10s┴%-6s┴%-12s┴%-12s┘\n",
		"────────────────────", "──────────", "──────", "────────────", "────────────")

	// Calculate success rate
	healthyCount := 0
	for _, item := range history {
		if item.Status == "healthy" {
			healthyCount++
		}
	}
	successRate := float64(healthyCount) / float64(len(history)) * 100

	fmt.Printf("\n📊 Success Rate: %.1f%% (%d/%d checks successful)\n", successRate, healthyCount, len(history))
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	
	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}