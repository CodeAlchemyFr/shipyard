package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/monitoring"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics [app-name]",
	Short: "Show detailed metrics for applications",
	Long: `Display detailed resource metrics for your deployed applications.

This command shows:
- CPU and memory usage over time
- Pod metrics breakdown
- Request rates and response times
- Historical trends

Examples:
  shipyard metrics                    # Show metrics for all apps
  shipyard metrics my-app             # Show metrics for specific app
  shipyard metrics --period 1h       # Show metrics for the last hour
  shipyard metrics --format table    # Display in table format`,
	Run: func(cmd *cobra.Command, args []string) {
		var appName string
		if len(args) > 0 {
			appName = args[0]
		}

		period, _ := cmd.Flags().GetDuration("period")
		format, _ := cmd.Flags().GetString("format")

		if err := runMetrics(appName, period, format); err != nil {
			log.Fatalf("Metrics command failed: %v", err)
		}
	},
}

func init() {
	metricsCmd.Flags().DurationP("period", "p", time.Hour, "Time period to show metrics for")
	metricsCmd.Flags().StringP("format", "f", "table", "Output format (table, json, csv)")
}

func runMetrics(appName string, period time.Duration, format string) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	// Get metrics for the specified period
	metrics, err := getMetricsForPeriod(collector, appName, period)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// Display metrics based on format
	switch format {
	case "table":
		displayMetricsTable(metrics, appName, period)
	case "json":
		displayMetricsJSON(metrics)
	case "csv":
		displayMetricsCSV(metrics)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// MetricsSummary represents aggregated metrics for an app
type MetricsSummary struct {
	AppName          string
	AvgCPU           float64
	MaxCPU           float64
	AvgMemory        float64
	MaxMemory        float64
	PodCount         int
	RequestsPerSec   float64
	AvgResponseTime  float64
	DataPoints       int
	Period           time.Duration
}

func getMetricsForPeriod(collector *monitoring.Collector, appName string, period time.Duration) ([]MetricsSummary, error) {
	// This is a simplified implementation - in a real scenario, you'd query the database
	// for historical metrics within the specified period
	
	// For now, return sample data
	summaries := []MetricsSummary{
		{
			AppName:         "web-app",
			AvgCPU:          42.5,
			MaxCPU:          78.2,
			AvgMemory:       256.0,
			MaxMemory:       412.8,
			PodCount:        3,
			RequestsPerSec:  125.5,
			AvgResponseTime: 89.3,
			DataPoints:      120,
			Period:          period,
		},
		{
			AppName:         "api-service",
			AvgCPU:          65.1,
			MaxCPU:          89.7,
			AvgMemory:       512.0,
			MaxMemory:       724.1,
			PodCount:        2,
			RequestsPerSec:  89.2,
			AvgResponseTime: 156.8,
			DataPoints:      118,
			Period:          period,
		},
	}

	// Filter by app name if specified
	if appName != "" {
		var filtered []MetricsSummary
		for _, summary := range summaries {
			if summary.AppName == appName {
				filtered = append(filtered, summary)
			}
		}
		return filtered, nil
	}

	return summaries, nil
}

func displayMetricsTable(metrics []MetricsSummary, appName string, period time.Duration) {
	if len(metrics) == 0 {
		fmt.Println("📊 No metrics found for the specified criteria")
		return
	}

	title := fmt.Sprintf("📊 Application Metrics (Last %v)", period)
	if appName != "" {
		title = fmt.Sprintf("📊 Metrics for %s (Last %v)", appName, period)
	}
	
	fmt.Println(title)
	fmt.Println("=" + fmt.Sprintf("%*s", len(title)-1, ""))
	fmt.Println()

	// Header
	fmt.Printf("┌%-15s┬%-10s┬%-10s┬%-12s┬%-12s┬%-8s┬%-10s┬%-12s┐\n", 
		"───────────────", "──────────", "──────────", "────────────", "────────────", "────────", "──────────", "────────────")
	fmt.Printf("│%-15s│%-10s│%-10s│%-12s│%-12s│%-8s│%-10s│%-12s│\n",
		"APP NAME", "AVG CPU%", "MAX CPU%", "AVG MEM(MB)", "MAX MEM(MB)", "PODS", "REQ/SEC", "AVG RESP(ms)")
	fmt.Printf("├%-15s┼%-10s┼%-10s┼%-12s┼%-12s┼%-8s┼%-10s┼%-12s┤\n",
		"───────────────", "──────────", "──────────", "────────────", "────────────", "────────", "──────────", "────────────")

	// Data rows
	for _, metric := range metrics {
		fmt.Printf("│%-15s│%-10.1f│%-10.1f│%-12.1f│%-12.1f│%-8d│%-10.1f│%-12.1f│\n",
			truncateString(metric.AppName, 15),
			metric.AvgCPU,
			metric.MaxCPU,
			metric.AvgMemory,
			metric.MaxMemory,
			metric.PodCount,
			metric.RequestsPerSec,
			metric.AvgResponseTime,
		)
	}

	fmt.Printf("└%-15s┴%-10s┴%-10s┴%-12s┴%-12s┴%-8s┴%-10s┴%-12s┘\n",
		"───────────────", "──────────", "──────────", "────────────", "────────────", "────────", "──────────", "────────────")
	
	fmt.Printf("\n💡 Tip: Use 'shipyard monitor' for real-time monitoring\n")
}

func displayMetricsJSON(metrics []MetricsSummary) {
	fmt.Printf("{\n  \"metrics\": [\n")
	for i, metric := range metrics {
		comma := ","
		if i == len(metrics)-1 {
			comma = ""
		}
		fmt.Printf(`    {
      "app_name": "%s",
      "avg_cpu_percent": %.1f,
      "max_cpu_percent": %.1f,
      "avg_memory_mb": %.1f,
      "max_memory_mb": %.1f,
      "pod_count": %d,
      "requests_per_second": %.1f,
      "avg_response_time_ms": %.1f,
      "data_points": %d,
      "period_seconds": %.0f
    }%s
`, metric.AppName, metric.AvgCPU, metric.MaxCPU, metric.AvgMemory, metric.MaxMemory,
			metric.PodCount, metric.RequestsPerSec, metric.AvgResponseTime, metric.DataPoints, metric.Period.Seconds(), comma)
	}
	fmt.Printf("  ]\n}\n")
}

func displayMetricsCSV(metrics []MetricsSummary) {
	fmt.Println("app_name,avg_cpu_percent,max_cpu_percent,avg_memory_mb,max_memory_mb,pod_count,requests_per_second,avg_response_time_ms,data_points,period_seconds")
	for _, metric := range metrics {
		fmt.Printf("%s,%.1f,%.1f,%.1f,%.1f,%d,%.1f,%.1f,%d,%.0f\n",
			metric.AppName, metric.AvgCPU, metric.MaxCPU, metric.AvgMemory, metric.MaxMemory,
			metric.PodCount, metric.RequestsPerSec, metric.AvgResponseTime, metric.DataPoints, metric.Period.Seconds())
	}
}