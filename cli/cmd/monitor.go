package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/monitoring"
	corev1 "k8s.io/api/core/v1"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor [app-name]",
	Short: "Monitor applications in real-time",
	Long: `Monitor the status, metrics, and health of your deployed applications in real-time.

This command provides a live dashboard showing:
- Application status and health
- CPU and memory usage
- Pod counts and readiness
- Active alerts
- Recent events

Examples:
  shipyard monitor                    # Monitor all applications
  shipyard monitor my-app             # Monitor specific application
  shipyard monitor --interval 10s    # Custom refresh interval
  shipyard monitor --alerts-only     # Show only applications with alerts`,
	Run: func(cmd *cobra.Command, args []string) {
		var appName string
		if len(args) > 0 {
			appName = args[0]
		}

		interval, _ := cmd.Flags().GetDuration("interval")
		alertsOnly, _ := cmd.Flags().GetBool("alerts-only")
		compact, _ := cmd.Flags().GetBool("compact")

		if err := runMonitor(appName, interval, alertsOnly, compact); err != nil {
			log.Fatalf("Monitor failed: %v", err)
		}
	},
}

func init() {
	monitorCmd.Flags().DurationP("interval", "i", 5*time.Second, "Refresh interval")
	monitorCmd.Flags().Bool("alerts-only", false, "Show only applications with active alerts")
	monitorCmd.Flags().Bool("compact", false, "Use compact display mode")
}

func runMonitor(appName string, interval time.Duration, alertsOnly, compact bool) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	// Initialize monitor display
	monitor := NewMonitorDisplay(collector, compact)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n👋 Stopping monitor...")
		cancel()
	}()

	fmt.Printf("🔍 Starting Shipyard Monitor (refresh: %v)\n", interval)
	fmt.Println("Press Ctrl+C to stop, 'h' for help")
	fmt.Println("")

	// Start monitoring loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial display
	if err := monitor.Update(appName, alertsOnly); err != nil {
		return fmt.Errorf("failed to update monitor: %w", err)
	}
	monitor.Render()

	// Monitoring loop
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := monitor.Update(appName, alertsOnly); err != nil {
				fmt.Printf("Warning: failed to update monitor: %v\n", err)
				continue
			}
			monitor.Render()
		}
	}
}

// MonitorDisplay handles the display of monitoring information
type MonitorDisplay struct {
	collector    *monitoring.Collector
	compact      bool
	lastUpdate   time.Time
	appStatuses  []monitoring.AppStatus
	clusterStatus monitoring.ClusterStatus
}

// NewMonitorDisplay creates a new monitor display
func NewMonitorDisplay(collector *monitoring.Collector, compact bool) *MonitorDisplay {
	return &MonitorDisplay{
		collector: collector,
		compact:   compact,
	}
}

// Update refreshes the monitoring data
func (m *MonitorDisplay) Update(appName string, alertsOnly bool) error {
	m.lastUpdate = time.Now()

	// Collect latest metrics
	if err := m.collector.CollectMetrics(appName); err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Get application statuses
	appStatuses, err := m.getAppStatuses(appName, alertsOnly)
	if err != nil {
		return fmt.Errorf("failed to get app statuses: %w", err)
	}
	m.appStatuses = appStatuses

	// Get cluster status
	clusterStatus, err := m.getClusterStatus()
	if err != nil {
		return fmt.Errorf("failed to get cluster status: %w", err)
	}
	m.clusterStatus = clusterStatus

	return nil
}

// Render displays the monitoring dashboard
func (m *MonitorDisplay) Render() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Header
	m.renderHeader()

	// Application table
	if len(m.appStatuses) > 0 {
		m.renderAppTable()
	} else {
		fmt.Println("📱 No applications found or matching criteria")
	}

	// Footer with cluster info
	m.renderFooter()
}

func (m *MonitorDisplay) renderHeader() {
	title := "┌─ Shipyard Monitor ─────────────────────────────────────────────────────┐"
	
	refreshInfo := fmt.Sprintf("Last updated: %s | Press 'q' to quit | Press 'h' for help",
		m.lastUpdate.Format("15:04:05"))
	
	// Pad to match title width
	padding := len(title) - len(refreshInfo) - 4
	if padding < 0 {
		padding = 0
	}
	
	fmt.Printf("%s\n", title)
	fmt.Printf("│ %s%s │\n", refreshInfo, strings.Repeat(" ", padding))
	fmt.Println("├────────────────────────────────────────────────────────────────────────┤")
}

func (m *MonitorDisplay) renderAppTable() {
	if m.compact {
		m.renderCompactTable()
	} else {
		m.renderDetailedTable()
	}
}

func (m *MonitorDisplay) renderDetailedTable() {
	// Table header
	fmt.Printf("│ %-15s │ %-8s │ %-6s │ %-8s │ %-9s │ %-12s │ %-6s │\n",
		"APP NAME", "STATUS", "CPU", "MEMORY", "REPLICAS", "LAST DEPLOY", "ALERTS")
	fmt.Println("├─────────────────┼──────────┼────────┼──────────┼───────────┼──────────────┼────────┤")

	for _, app := range m.appStatuses {
		cpuStr := "N/A"
		memStr := "N/A"
		if app.Metrics != nil {
			cpuStr = fmt.Sprintf("%.1f%%", app.Metrics.CPUPercent)
			memStr = fmt.Sprintf("%.0fMB", float64(app.Metrics.MemoryBytes)/(1024*1024))
		}

		alertStr := fmt.Sprintf("%d", app.ActiveAlerts)
		if app.ActiveAlerts > 0 {
			alertStr = fmt.Sprintf("⚠️ %d", app.ActiveAlerts)
		}

		fmt.Printf("│ %-15s │ %s %-6s │ %-6s │ %-8s │ %-9s │ %-12s │ %-6s │\n",
			truncateString(app.AppName, 15),
			app.StatusIcon,
			truncateString(app.Status, 6),
			cpuStr,
			memStr,
			app.Replicas,
			app.LastDeploy,
			alertStr)
	}
}

func (m *MonitorDisplay) renderCompactTable() {
	// Compact view with fewer columns
	fmt.Printf("│ %-20s │ %-8s │ %-12s │ %-8s │\n",
		"APP NAME", "STATUS", "RESOURCES", "ALERTS")
	fmt.Println("├──────────────────────┼──────────┼──────────────┼──────────┤")

	for _, app := range m.appStatuses {
		resourceStr := "N/A"
		if app.Metrics != nil {
			resourceStr = fmt.Sprintf("%.0f%% / %.0fMB",
				app.Metrics.CPUPercent,
				float64(app.Metrics.MemoryBytes)/(1024*1024))
		}

		alertStr := fmt.Sprintf("%d", app.ActiveAlerts)
		if app.ActiveAlerts > 0 {
			alertStr = fmt.Sprintf("⚠️ %d", app.ActiveAlerts)
		}

		fmt.Printf("│ %-20s │ %s %-6s │ %-12s │ %-8s │\n",
			truncateString(app.AppName, 20),
			app.StatusIcon,
			truncateString(app.Status, 6),
			resourceStr,
			alertStr)
	}
}

func (m *MonitorDisplay) renderFooter() {
	fmt.Println("├────────────────────────────────────────────────────────────────────────┤")
	
	healthIcon := "✅"
	if !m.clusterStatus.Healthy {
		healthIcon = "❌"
	}

	clusterInfo := fmt.Sprintf("Cluster: %s %s | Nodes: %d/%d | Total Pods: %d",
		healthIcon,
		"Healthy",
		m.clusterStatus.NodesReady,
		m.clusterStatus.NodesTotal,
		m.clusterStatus.PodsTotal)

	if m.clusterStatus.AlertsTotal > 0 {
		clusterInfo += fmt.Sprintf(" | Alerts: %d active", m.clusterStatus.AlertsTotal)
	}

	// Pad to match table width
	padding := 72 - len(clusterInfo)
	if padding < 0 {
		padding = 0
	}

	fmt.Printf("│ %s%s │\n", clusterInfo, strings.Repeat(" ", padding))
	fmt.Println("└────────────────────────────────────────────────────────────────────────┘")
}

// Helper methods

func (m *MonitorDisplay) getAppStatuses(appName string, alertsOnly bool) ([]monitoring.AppStatus, error) {
	// Get apps from database
	apps, err := m.collector.GetAppsToMonitor(appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}

	var statuses []monitoring.AppStatus
	for _, app := range apps {
		status, err := m.getAppStatus(app, alertsOnly)
		if err != nil {
			// Continue with other apps if one fails
			fmt.Printf("Warning: failed to get status for %s: %v\n", app.Name, err)
			continue
		}
		
		// Filter by alerts if requested
		if alertsOnly && status.ActiveAlerts == 0 {
			continue
		}
		
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (m *MonitorDisplay) getAppStatus(app monitoring.App, alertsOnly bool) (monitoring.AppStatus, error) {
	status := monitoring.AppStatus{
		AppName: app.Name,
		Status:  "unknown",
		StatusIcon: "⚪",
		Replicas: "0/0",
		LastDeploy: "unknown",
		ActiveAlerts: 0,
	}

	// Get deployment info from Kubernetes
	deployment, err := m.collector.GetK8sClient().GetDeployment(app.Name)
	if err != nil {
		// App might not be deployed yet
		status.Status = "not-deployed"
		status.StatusIcon = "⚫"
		return status, nil
	}

	// Calculate replicas
	desired := int32(0)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}
	ready := deployment.Status.ReadyReplicas
	status.Replicas = fmt.Sprintf("%d/%d", ready, desired)

	// Determine status based on replicas
	if ready == desired && desired > 0 {
		status.Status = "healthy"
		status.StatusIcon = "🟢"
	} else if ready > 0 {
		status.Status = "degraded"
		status.StatusIcon = "🟡"
	} else {
		status.Status = "failed"
		status.StatusIcon = "🔴"
	}

	// Get last deployment time
	if deployment.CreationTimestamp.Time.IsZero() {
		status.LastDeploy = "unknown"
	} else {
		timeSince := time.Since(deployment.CreationTimestamp.Time)
		status.LastDeploy = formatTimeSince(timeSince)
	}

	// Get metrics
	metrics, err := m.getAppMetrics(app)
	if err == nil {
		status.Metrics = metrics
	}

	// Get active alerts count
	alertCount, err := m.getActiveAlertsCount(app.ID)
	if err == nil {
		status.ActiveAlerts = alertCount
		
		// Update status based on alerts
		if alertCount > 0 && status.Status == "healthy" {
			status.Status = "warning"
			status.StatusIcon = "🟡"
		}
	}

	return status, nil
}

func (m *MonitorDisplay) getAppMetrics(app monitoring.App) (*monitoring.AppMetrics, error) {
	// Get pod metrics from Kubernetes
	podMetrics, err := m.collector.GetK8sClient().GetPodMetrics(app.Name)
	if err != nil {
		return nil, err
	}

	if len(podMetrics) == 0 {
		return nil, fmt.Errorf("no pod metrics available")
	}

	// Get pods for counting
	pods, err := m.collector.GetK8sClient().GetPods(app.Name)
	if err != nil {
		return nil, err
	}

	// Calculate aggregate metrics
	totalCPU := int64(0)
	totalMemory := int64(0)
	readyPods := 0

	for _, podMetric := range podMetrics {
		if len(podMetric.Containers) > 0 {
			totalCPU += podMetric.Containers[0].Usage.Cpu().MilliValue()
			totalMemory += podMetric.Containers[0].Usage.Memory().Value()
		}
	}

	for _, pod := range pods {
		if m.isPodReady(pod) {
			readyPods++
		}
	}

	// Convert CPU from millicores to percentage (simplified)
	// This is a rough estimation - in production you'd use resource requests/limits
	cpuPercent := float64(totalCPU) / 10.0 // Rough conversion
	if cpuPercent > 100 {
		cpuPercent = 100
	}

	return &monitoring.AppMetrics{
		CPUPercent:     cpuPercent,
		MemoryBytes:    totalMemory,
		PodCount:       len(pods),
		PodReady:       readyPods,
		RequestsPerSec: 0, // Would need additional metrics collection
		LastUpdated:    time.Now(),
	}, nil
}

func (m *MonitorDisplay) isPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func (m *MonitorDisplay) getActiveAlertsCount(appID int64) (int, error) {
	// Query database for active alerts
	count, err := m.collector.GetActiveAlertsCount(appID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func formatTimeSince(d time.Duration) string {
	if d < time.Minute {
		return "now"
	} else if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func (m *MonitorDisplay) getClusterStatus() (monitoring.ClusterStatus, error) {
	// Get cluster info from Kubernetes
	clusterInfo, err := m.collector.GetK8sClient().GetClusterInfo()
	if err != nil {
		return monitoring.ClusterStatus{
			Healthy:     false,
			NodesTotal:  0,
			NodesReady:  0,
			PodsTotal:   0,
			AlertsTotal: 0,
			Message:     "Unable to connect to cluster",
		}, nil
	}

	// Get total active alerts
	alertsTotal, err := m.getTotalActiveAlerts()
	if err != nil {
		alertsTotal = 0
	}

	nodesTotal := clusterInfo["nodes_total"].(int)
	nodesReady := clusterInfo["nodes_ready"].(int)
	podsTotal := clusterInfo["pods_total"].(int)

	healthy := nodesReady == nodesTotal && alertsTotal == 0
	message := "All systems operational"
	if !healthy {
		if nodesReady < nodesTotal {
			message = fmt.Sprintf("%d/%d nodes ready", nodesReady, nodesTotal)
		}
		if alertsTotal > 0 {
			if message != "All systems operational" {
				message += fmt.Sprintf(", %d alerts", alertsTotal)
			} else {
				message = fmt.Sprintf("%d active alerts", alertsTotal)
			}
		}
	}

	return monitoring.ClusterStatus{
		Healthy:     healthy,
		NodesTotal:  nodesTotal,
		NodesReady:  nodesReady,
		PodsTotal:   podsTotal,
		AlertsTotal: alertsTotal,
		Message:     message,
	}, nil
}

func (m *MonitorDisplay) getTotalActiveAlerts() (int, error) {
	// Query database for total active alerts across all apps
	count, err := m.collector.GetTotalActiveAlerts()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}