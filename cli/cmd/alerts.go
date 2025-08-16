package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/monitoring"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts [command]",
	Short: "Manage application alerts",
	Long: `Manage alerts for your deployed applications.

Available commands:
  list      List active alerts
  history   Show alert history
  resolve   Resolve an alert
  config    Configure alert thresholds

Examples:
  shipyard alerts list                # List all active alerts
  shipyard alerts list my-app         # List alerts for specific app
  shipyard alerts history --period 1d # Show alerts from last day
  shipyard alerts resolve 123         # Resolve alert with ID 123
  shipyard alerts config my-app       # Configure alerts for app`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Default to list command
			if err := runAlertsList("", false); err != nil {
				log.Fatalf("Alerts command failed: %v", err)
			}
			return
		}

		subCommand := args[0]
		switch subCommand {
		case "list":
			appName := ""
			if len(args) > 1 {
				appName = args[1]
			}
			activeOnly, _ := cmd.Flags().GetBool("active")
			if err := runAlertsList(appName, activeOnly); err != nil {
				log.Fatalf("Alerts list failed: %v", err)
			}
		case "history":
			appName := ""
			if len(args) > 1 {
				appName = args[1]
			}
			period, _ := cmd.Flags().GetDuration("period")
			if err := runAlertsHistory(appName, period); err != nil {
				log.Fatalf("Alerts history failed: %v", err)
			}
		case "resolve":
			if len(args) < 2 {
				fmt.Println("Error: Alert ID required")
				cmd.Help()
				return
			}
			alertID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid alert ID: %v\n", err)
				return
			}
			if err := runAlertsResolve(alertID); err != nil {
				log.Fatalf("Resolve alert failed: %v", err)
			}
		case "config":
			appName := ""
			if len(args) > 1 {
				appName = args[1]
			}
			if err := runAlertsConfig(appName); err != nil {
				log.Fatalf("Configure alerts failed: %v", err)
			}
		default:
			fmt.Printf("Unknown command: %s\n", subCommand)
			cmd.Help()
		}
	},
}

func init() {
	alertsCmd.Flags().BoolP("active", "a", false, "Show only active alerts")
	alertsCmd.Flags().DurationP("period", "p", 24*time.Hour, "Time period for history")
}

// AlertInfo represents alert information for display
type AlertInfo struct {
	ID           int64
	AppName      string
	Type         string
	Severity     string
	Status       string
	Message      string
	Threshold    float64
	CurrentValue float64
	CreatedAt    time.Time
	ResolvedAt   *time.Time
	Duration     time.Duration
}

func runAlertsList(appName string, activeOnly bool) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	alerts, err := getAlerts(collector, appName, activeOnly)
	if err != nil {
		return fmt.Errorf("failed to get alerts: %w", err)
	}

	displayAlertsTable(alerts, appName, activeOnly)
	return nil
}

func runAlertsHistory(appName string, period time.Duration) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	alerts, err := getAlertsHistory(collector, appName, period)
	if err != nil {
		return fmt.Errorf("failed to get alerts history: %w", err)
	}

	displayAlertsHistory(alerts, appName, period)
	return nil
}

func runAlertsResolve(alertID int64) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	if err := resolveAlert(collector, alertID); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	fmt.Printf("✅ Alert %d resolved successfully\n", alertID)
	return nil
}

func runAlertsConfig(appName string) error {
	if appName == "" {
		return fmt.Errorf("app name required for configuration")
	}

	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	config, err := getAlertConfig(collector, appName)
	if err != nil {
		return fmt.Errorf("failed to get alert configuration: %w", err)
	}

	displayAlertConfig(config, appName)
	return nil
}

func getAlerts(collector *monitoring.Collector, appName string, activeOnly bool) ([]AlertInfo, error) {
	// This would normally query the database for alerts
	// For now, return sample data
	
	now := time.Now()
	alerts := []AlertInfo{
		{
			ID:           1,
			AppName:      "web-app",
			Type:         "cpu_high",
			Severity:     "warning",
			Status:       "active",
			Message:      "CPU usage 85.2% exceeds threshold 80.0%",
			Threshold:    80.0,
			CurrentValue: 85.2,
			CreatedAt:    now.Add(-2 * time.Hour),
			Duration:     2 * time.Hour,
		},
		{
			ID:           2,
			AppName:      "api-service",
			Type:         "memory_high",
			Severity:     "critical",
			Status:       "active",
			Message:      "Memory usage 612MB exceeds threshold 500MB",
			Threshold:    500.0,
			CurrentValue: 612.0,
			CreatedAt:    now.Add(-45 * time.Minute),
			Duration:     45 * time.Minute,
		},
		{
			ID:           3,
			AppName:      "api-service",
			Type:         "response_time_high",
			Severity:     "warning",
			Status:       "active",
			Message:      "Average response time 1250ms exceeds threshold 1000ms",
			Threshold:    1000.0,
			CurrentValue: 1250.0,
			CreatedAt:    now.Add(-30 * time.Minute),
			Duration:     30 * time.Minute,
		},
		{
			ID:           4,
			AppName:      "worker",
			Type:         "pod_restart",
			Severity:     "warning",
			Status:       "resolved",
			Message:      "Pod restarted 3 times in the last hour",
			Threshold:    2.0,
			CurrentValue: 3.0,
			CreatedAt:    now.Add(-3 * time.Hour),
			ResolvedAt:   &[]time.Time{now.Add(-1 * time.Hour)}[0],
			Duration:     2 * time.Hour,
		},
	}

	// Filter by app name if specified
	if appName != "" {
		var filtered []AlertInfo
		for _, alert := range alerts {
			if alert.AppName == appName {
				filtered = append(filtered, alert)
			}
		}
		alerts = filtered
	}

	// Filter by status if activeOnly
	if activeOnly {
		var filtered []AlertInfo
		for _, alert := range alerts {
			if alert.Status == "active" {
				filtered = append(filtered, alert)
			}
		}
		alerts = filtered
	}

	return alerts, nil
}

func getAlertsHistory(collector *monitoring.Collector, appName string, period time.Duration) ([]AlertInfo, error) {
	// This would normally query the database for historical alerts
	// For now, return sample data by calling getAlerts and filtering by time
	
	alerts, err := getAlerts(collector, appName, false)
	if err != nil {
		return nil, err
	}

	// Filter by period
	cutoff := time.Now().Add(-period)
	var filtered []AlertInfo
	for _, alert := range alerts {
		if alert.CreatedAt.After(cutoff) {
			filtered = append(filtered, alert)
		}
	}

	return filtered, nil
}

func resolveAlert(collector *monitoring.Collector, alertID int64) error {
	// This would normally update the database to resolve the alert
	// For now, just simulate success
	fmt.Printf("Resolving alert ID %d...\n", alertID)
	time.Sleep(500 * time.Millisecond) // Simulate processing
	return nil
}

// AlertConfig represents alert configuration
type AlertConfig struct {
	AppName                string
	CPUThreshold           float64
	MemoryThreshold        float64
	ResponseTimeThreshold  int
	ErrorRateThreshold     float64
	HealthCheckEnabled     bool
	HealthCheckInterval    int
	NotificationEnabled    bool
	NotificationChannels   []string
}

func getAlertConfig(collector *monitoring.Collector, appName string) (*AlertConfig, error) {
	// This would normally query the database for alert configuration
	// For now, return sample data
	
	return &AlertConfig{
		AppName:                appName,
		CPUThreshold:           80.0,
		MemoryThreshold:        85.0,
		ResponseTimeThreshold:  1000,
		ErrorRateThreshold:     5.0,
		HealthCheckEnabled:     true,
		HealthCheckInterval:    30,
		NotificationEnabled:    false,
		NotificationChannels:   []string{"email", "slack"},
	}, nil
}

func displayAlertsTable(alerts []AlertInfo, appName string, activeOnly bool) {
	if len(alerts) == 0 {
		status := "alerts"
		if activeOnly {
			status = "active alerts"
		}
		fmt.Printf("🚨 No %s found", status)
		if appName != "" {
			fmt.Printf(" for %s", appName)
		}
		fmt.Println()
		return
	}

	title := "🚨 Application Alerts"
	if appName != "" {
		title = fmt.Sprintf("🚨 Alerts for %s", appName)
	}
	if activeOnly {
		title += " (Active Only)"
	}
	
	fmt.Println(title)
	fmt.Println("=" + fmt.Sprintf("%*s", len(title)-1, ""))
	fmt.Println()

	// Header
	fmt.Printf("┌%-4s┬%-12s┬%-15s┬%-10s┬%-8s┬%-12s┬%-40s┐\n",
		"────", "────────────", "───────────────", "──────────", "────────", "────────────", "────────────────────────────────────────")
	fmt.Printf("│%-4s│%-12s│%-15s│%-10s│%-8s│%-12s│%-40s│\n",
		"ID", "APP", "TYPE", "SEVERITY", "STATUS", "DURATION", "MESSAGE")
	fmt.Printf("├%-4s┼%-12s┼%-15s┼%-10s┼%-8s┼%-12s┼%-40s┤\n",
		"────", "────────────", "───────────────", "──────────", "────────", "────────────", "────────────────────────────────────────")

	// Data rows
	for _, alert := range alerts {
		severityIcon := "⚠️"
		if alert.Severity == "critical" {
			severityIcon = "🔴"
		} else if alert.Severity == "info" {
			severityIcon = "🔵"
		}

		statusIcon := "🟡"
		if alert.Status == "resolved" {
			statusIcon = "✅"
		}

		durationStr := formatDuration(alert.Duration)

		fmt.Printf("│%-4d│%-12s│%-15s│%s %-8s│%s %-6s│%-12s│%-40s│\n",
			alert.ID,
			truncateString(alert.AppName, 12),
			truncateString(alert.Type, 15),
			severityIcon,
			truncateString(alert.Severity, 7),
			statusIcon,
			truncateString(alert.Status, 5),
			durationStr,
			truncateString(alert.Message, 40),
		)
	}

	fmt.Printf("└%-4s┴%-12s┴%-15s┴%-10s┴%-8s┴%-12s┴%-40s┘\n",
		"────", "────────────", "───────────────", "──────────", "────────", "────────────", "────────────────────────────────────────")
	
	// Summary
	activeCount := 0
	criticalCount := 0
	for _, alert := range alerts {
		if alert.Status == "active" {
			activeCount++
			if alert.Severity == "critical" {
				criticalCount++
			}
		}
	}

	fmt.Printf("\n📊 Summary: %d total alerts", len(alerts))
	if activeCount > 0 {
		fmt.Printf(", %d active", activeCount)
		if criticalCount > 0 {
			fmt.Printf(" (%d critical)", criticalCount)
		}
	}
	fmt.Println()
	fmt.Printf("💡 Tip: Use 'shipyard alerts resolve <id>' to resolve alerts\n")
}

func displayAlertsHistory(alerts []AlertInfo, appName string, period time.Duration) {
	if len(alerts) == 0 {
		fmt.Printf("🚨 No alerts found in the last %v", period)
		if appName != "" {
			fmt.Printf(" for %s", appName)
		}
		fmt.Println()
		return
	}

	title := fmt.Sprintf("🚨 Alert History (Last %v)", period)
	if appName != "" {
		title = fmt.Sprintf("🚨 Alert History for %s (Last %v)", appName, period)
	}
	
	fmt.Println(title)
	fmt.Println("=" + fmt.Sprintf("%*s", len(title)-1, ""))
	fmt.Println()

	// Group alerts by day
	dayGroups := make(map[string][]AlertInfo)
	for _, alert := range alerts {
		day := alert.CreatedAt.Format("2006-01-02")
		dayGroups[day] = append(dayGroups[day], alert)
	}

	// Display by day
	for day, dayAlerts := range dayGroups {
		fmt.Printf("📅 %s (%d alerts)\n", day, len(dayAlerts))
		fmt.Println(strings.Repeat("─", 50))
		
		for _, alert := range dayAlerts {
			statusIcon := "🟡"
			if alert.Status == "resolved" {
				statusIcon = "✅"
			}
			
			severityIcon := "⚠️"
			if alert.Severity == "critical" {
				severityIcon = "🔴"
			}
			
			timeStr := alert.CreatedAt.Format("15:04")
			durationStr := formatDuration(alert.Duration)
			
			fmt.Printf("  %s %s [%s] %s (%s) - %s\n",
				timeStr,
				statusIcon,
				alert.AppName,
				severityIcon,
				durationStr,
				alert.Message,
			)
		}
		fmt.Println()
	}

	// Calculate statistics
	totalAlerts := len(alerts)
	resolvedAlerts := 0
	criticalAlerts := 0
	for _, alert := range alerts {
		if alert.Status == "resolved" {
			resolvedAlerts++
		}
		if alert.Severity == "critical" {
			criticalAlerts++
		}
	}

	fmt.Printf("📊 Summary: %d total alerts, %d resolved (%.1f%%), %d critical\n",
		totalAlerts, resolvedAlerts, float64(resolvedAlerts)/float64(totalAlerts)*100, criticalAlerts)
}

func displayAlertConfig(config *AlertConfig, appName string) {
	fmt.Printf("⚙️ Alert Configuration for %s\n", appName)
	fmt.Println("=" + fmt.Sprintf("%*s", len(appName)+23, ""))
	fmt.Println()

	fmt.Printf("Thresholds:\n")
	fmt.Printf("  CPU Usage:       %.1f%%\n", config.CPUThreshold)
	fmt.Printf("  Memory Usage:    %.1f%%\n", config.MemoryThreshold)
	fmt.Printf("  Response Time:   %dms\n", config.ResponseTimeThreshold)
	fmt.Printf("  Error Rate:      %.1f%%\n", config.ErrorRateThreshold)
	fmt.Println()

	fmt.Printf("Health Checks:\n")
	fmt.Printf("  Enabled:         %t\n", config.HealthCheckEnabled)
	fmt.Printf("  Interval:        %ds\n", config.HealthCheckInterval)
	fmt.Println()

	fmt.Printf("Notifications:\n")
	fmt.Printf("  Enabled:         %t\n", config.NotificationEnabled)
	fmt.Printf("  Channels:        %s\n", strings.Join(config.NotificationChannels, ", "))
	fmt.Println()

	fmt.Printf("💡 Use 'shipyard config edit %s' to modify these settings\n", appName)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		return fmt.Sprintf("%.1fd", d.Hours()/24)
	}
}