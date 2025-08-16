package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shipyard/cli/pkg/database"
	"github.com/shipyard/cli/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
)

// Collector handles metrics collection from Kubernetes
type Collector struct {
	db     *database.DB
	k8s    *k8s.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCollector creates a new metrics collector
func NewCollector() (*Collector, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize k8s client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		db:     db,
		k8s:    k8sClient,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close stops the collector and closes connections
func (c *Collector) Close() error {
	c.cancel()
	return c.db.Close()
}

// CollectMetrics collects metrics for all or specific applications
func (c *Collector) CollectMetrics(appName string) error {
	// Get applications to monitor
	apps, err := c.getAppsToMonitor(appName)
	if err != nil {
		return fmt.Errorf("failed to get apps: %w", err)
	}

	for _, app := range apps {
		if err := c.collectAppMetrics(app); err != nil {
			fmt.Printf("Warning: failed to collect metrics for %s: %v\n", app.Name, err)
		}
	}

	return nil
}

// collectAppMetrics collects all metrics for a specific application
func (c *Collector) collectAppMetrics(app App) error {
	// Collect pod metrics
	if err := c.collectPodMetrics(app); err != nil {
		return fmt.Errorf("failed to collect pod metrics: %w", err)
	}

	// Collect deployment metrics
	if err := c.collectDeploymentMetrics(app); err != nil {
		return fmt.Errorf("failed to collect deployment metrics: %w", err)
	}

	// Perform health checks
	if err := c.performHealthCheck(app); err != nil {
		fmt.Printf("Warning: health check failed for %s: %v\n", app.Name, err)
	}

	// Check alert conditions
	if err := c.checkAlerts(app); err != nil {
		fmt.Printf("Warning: alert check failed for %s: %v\n", app.Name, err)
	}

	return nil
}

// collectPodMetrics collects CPU and memory metrics from pods
func (c *Collector) collectPodMetrics(app App) error {
	// Get pods for the application
	pods, err := c.k8s.GetPods(app.Name)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	// Get pod metrics from metrics-server
	podMetrics, err := c.k8s.GetPodMetrics(app.Name)
	if err != nil {
		return fmt.Errorf("failed to get pod metrics: %w", err)
	}

	now := time.Now()

	// Store pod count
	if err := c.storeMetric(Metric{
		AppID:     app.ID,
		Type:      MetricTypePods,
		Value:     float64(len(pods)),
		Unit:      "count",
		Timestamp: now,
	}); err != nil {
		return fmt.Errorf("failed to store pod count: %w", err)
	}

	// Count ready pods
	readyPods := 0
	for _, pod := range pods {
		if c.isPodReady(pod) {
			readyPods++
		}
	}

	// Store metrics for each pod
	for _, podMetric := range podMetrics {
		podName := podMetric.Name
		
		// CPU metrics
		cpuUsage := podMetric.Containers[0].Usage.Cpu().MilliValue()
		if err := c.storeMetric(Metric{
			AppID:     app.ID,
			Type:      MetricTypeCPU,
			Value:     float64(cpuUsage),
			Unit:      "millicores",
			PodName:   podName,
			Timestamp: now,
		}); err != nil {
			continue
		}

		// Memory metrics
		memoryUsage := podMetric.Containers[0].Usage.Memory().Value()
		if err := c.storeMetric(Metric{
			AppID:     app.ID,
			Type:      MetricTypeMemory,
			Value:     float64(memoryUsage),
			Unit:      "bytes",
			PodName:   podName,
			Timestamp: now,
		}); err != nil {
			continue
		}
	}

	return nil
}

// collectDeploymentMetrics collects deployment-level metrics
func (c *Collector) collectDeploymentMetrics(app App) error {
	deployment, err := c.k8s.GetDeployment(app.Name)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	now := time.Now()

	// Store replica metrics
	if err := c.storeMetric(Metric{
		AppID:     app.ID,
		Type:      "replicas_desired",
		Value:     float64(*deployment.Spec.Replicas),
		Unit:      "count",
		Timestamp: now,
	}); err != nil {
		return err
	}

	if err := c.storeMetric(Metric{
		AppID:     app.ID,
		Type:      "replicas_ready",
		Value:     float64(deployment.Status.ReadyReplicas),
		Unit:      "count",
		Timestamp: now,
	}); err != nil {
		return err
	}

	return nil
}

// performHealthCheck performs HTTP health checks on application endpoints
func (c *Collector) performHealthCheck(app App) error {
	config, err := c.getMonitoringConfig(app.ID)
	if err != nil {
		return fmt.Errorf("failed to get monitoring config: %w", err)
	}

	if !config.Enabled {
		return nil
	}

	// Get service endpoint
	service, err := c.k8s.GetService(app.Name)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Build health check URL
	port := service.Spec.Ports[0].Port
	url := fmt.Sprintf("http://%s:%d%s", service.Spec.ClusterIP, port, config.HealthCheckPath)

	// Perform health check
	start := time.Now()
	resp, err := http.Get(url)
	responseTime := int(time.Since(start).Milliseconds())

	healthCheck := HealthCheck{
		AppID:        app.ID,
		Endpoint:     config.HealthCheckPath,
		Method:       "GET",
		ResponseTime: responseTime,
		CheckedAt:    time.Now(),
	}

	if err != nil {
		healthCheck.Status = HealthStatusError
		healthCheck.ErrorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		healthCheck.StatusCode = resp.StatusCode

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			healthCheck.Status = HealthStatusHealthy
		} else {
			healthCheck.Status = HealthStatusUnhealthy
		}
	}

	return c.storeHealthCheck(healthCheck)
}

// checkAlerts evaluates alert conditions and creates/resolves alerts
func (c *Collector) checkAlerts(app App) error {
	config, err := c.getMonitoringConfig(app.ID)
	if err != nil {
		return err
	}

	// Get recent metrics for alert evaluation
	metrics, err := c.getRecentMetrics(app.ID, time.Minute*5)
	if err != nil {
		return err
	}

	// Check CPU threshold
	if cpuMetric := c.getLatestMetric(metrics, MetricTypeCPU); cpuMetric != nil {
		cpuPercent := (cpuMetric.Value / 1000) * 100 // Convert millicores to percentage
		if cpuPercent > config.CPUThreshold {
			alert := Alert{
				AppID:        app.ID,
				Type:         "cpu_high",
				Threshold:    config.CPUThreshold,
				CurrentValue: cpuPercent,
				Severity:     AlertSeverityWarning,
				Status:       AlertStatusActive,
				Message:      fmt.Sprintf("CPU usage %.1f%% exceeds threshold %.1f%%", cpuPercent, config.CPUThreshold),
				CreatedAt:    time.Now(),
			}
			if cpuPercent > config.CPUThreshold*1.2 {
				alert.Severity = AlertSeverityCritical
			}
			c.createOrUpdateAlert(alert)
		} else {
			c.resolveAlert(app.ID, "cpu_high")
		}
	}

	// Check memory threshold
	if memMetric := c.getLatestMetric(metrics, MetricTypeMemory); memMetric != nil {
		// Note: This is a simplified calculation - in production you'd need pod resource limits
		memoryMB := memMetric.Value / (1024 * 1024)
		if memoryMB > 500 { // Example threshold
			alert := Alert{
				AppID:        app.ID,
				Type:         "memory_high",
				Threshold:    config.MemoryThreshold,
				CurrentValue: memoryMB,
				Severity:     AlertSeverityWarning,
				Status:       AlertStatusActive,
				Message:      fmt.Sprintf("Memory usage %.1fMB exceeds threshold", memoryMB),
				CreatedAt:    time.Now(),
			}
			c.createOrUpdateAlert(alert)
		} else {
			c.resolveAlert(app.ID, "memory_high")
		}
	}

	return nil
}

// Helper methods

func (c *Collector) isPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func (c *Collector) getLatestMetric(metrics []Metric, metricType MetricType) *Metric {
	for _, metric := range metrics {
		if metric.Type == metricType {
			return &metric
		}
	}
	return nil
}

// Database operations

func (c *Collector) storeMetric(metric Metric) error {
	query := `
		INSERT INTO metrics (app_id, metric_type, value, unit, pod_name, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := c.db.GetConnection().Exec(query,
		metric.AppID,
		string(metric.Type),
		metric.Value,
		metric.Unit,
		metric.PodName,
		metric.Timestamp,
	)
	return err
}

func (c *Collector) storeHealthCheck(hc HealthCheck) error {
	query := `
		INSERT INTO health_checks (app_id, endpoint, method, status, status_code, response_time, error_message, checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := c.db.GetConnection().Exec(query,
		hc.AppID,
		hc.Endpoint,
		hc.Method,
		string(hc.Status),
		hc.StatusCode,
		hc.ResponseTime,
		hc.ErrorMessage,
		hc.CheckedAt,
	)
	return err
}

func (c *Collector) createOrUpdateAlert(alert Alert) error {
	// Check if alert already exists
	var existingID int64
	query := `SELECT id FROM alerts WHERE app_id = ? AND alert_type = ? AND status = 'active'`
	err := c.db.GetConnection().QueryRow(query, alert.AppID, alert.Type).Scan(&existingID)

	if err == nil {
		// Update existing alert
		updateQuery := `
			UPDATE alerts 
			SET current_value = ?, message = ?, severity = ?
			WHERE id = ?`
		_, err = c.db.GetConnection().Exec(updateQuery,
			alert.CurrentValue,
			alert.Message,
			string(alert.Severity),
			existingID,
		)
		return err
	}

	// Create new alert
	insertQuery := `
		INSERT INTO alerts (app_id, alert_type, threshold, current_value, severity, status, message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = c.db.GetConnection().Exec(insertQuery,
		alert.AppID,
		alert.Type,
		alert.Threshold,
		alert.CurrentValue,
		string(alert.Severity),
		string(alert.Status),
		alert.Message,
		alert.CreatedAt,
	)
	return err
}

func (c *Collector) resolveAlert(appID int64, alertType string) error {
	query := `
		UPDATE alerts 
		SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP
		WHERE app_id = ? AND alert_type = ? AND status = 'active'`

	_, err := c.db.GetConnection().Exec(query, appID, alertType)
	return err
}

func (c *Collector) getRecentMetrics(appID int64, duration time.Duration) ([]Metric, error) {
	query := `
		SELECT app_id, metric_type, value, unit, pod_name, timestamp
		FROM metrics
		WHERE app_id = ? AND timestamp > datetime('now', '-' || ? || ' seconds')
		ORDER BY timestamp DESC`

	rows, err := c.db.GetConnection().Query(query, appID, int(duration.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metric Metric
		var metricType string
		err := rows.Scan(
			&metric.AppID,
			&metricType,
			&metric.Value,
			&metric.Unit,
			&metric.PodName,
			&metric.Timestamp,
		)
		if err != nil {
			continue
		}
		metric.Type = MetricType(metricType)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (c *Collector) getMonitoringConfig(appID int64) (*MonitoringConfig, error) {
	query := `
		SELECT id, app_id, enabled, health_check_path, health_check_interval, health_check_timeout,
		       metrics_enabled, metrics_path, metrics_port, retention_days,
		       cpu_threshold, memory_threshold, error_rate_threshold, response_time_threshold,
		       created_at, updated_at
		FROM monitoring_config
		WHERE app_id = ?`

	var config MonitoringConfig
	err := c.db.GetConnection().QueryRow(query, appID).Scan(
		&config.ID,
		&config.AppID,
		&config.Enabled,
		&config.HealthCheckPath,
		&config.HealthCheckInterval,
		&config.HealthCheckTimeout,
		&config.MetricsEnabled,
		&config.MetricsPath,
		&config.MetricsPort,
		&config.RetentionDays,
		&config.CPUThreshold,
		&config.MemoryThreshold,
		&config.ErrorRateThreshold,
		&config.ResponseTimeThreshold,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		// Create default config if not exists
		defaultConfig := &MonitoringConfig{
			AppID:                  appID,
			Enabled:                true,
			HealthCheckPath:        "/health",
			HealthCheckInterval:    30,
			HealthCheckTimeout:     5,
			MetricsEnabled:         true,
			MetricsPath:            "/metrics",
			MetricsPort:            9090,
			RetentionDays:          7,
			CPUThreshold:           80.0,
			MemoryThreshold:        85.0,
			ErrorRateThreshold:     5.0,
			ResponseTimeThreshold:  1000,
		}

		if err := c.createMonitoringConfig(defaultConfig); err != nil {
			return nil, err
		}
		return defaultConfig, nil
	}

	return &config, nil
}

func (c *Collector) createMonitoringConfig(config *MonitoringConfig) error {
	query := `
		INSERT INTO monitoring_config 
		(app_id, enabled, health_check_path, health_check_interval, health_check_timeout,
		 metrics_enabled, metrics_path, metrics_port, retention_days,
		 cpu_threshold, memory_threshold, error_rate_threshold, response_time_threshold)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := c.db.GetConnection().Exec(query,
		config.AppID,
		config.Enabled,
		config.HealthCheckPath,
		config.HealthCheckInterval,
		config.HealthCheckTimeout,
		config.MetricsEnabled,
		config.MetricsPath,
		config.MetricsPort,
		config.RetentionDays,
		config.CPUThreshold,
		config.MemoryThreshold,
		config.ErrorRateThreshold,
		config.ResponseTimeThreshold,
	)
	return err
}

// App represents an application for monitoring
type App struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// GetAppsToMonitor returns apps to monitor (exported method)
func (c *Collector) GetAppsToMonitor(appName string) ([]App, error) {
	return c.getAppsToMonitor(appName)
}

// GetK8sClient returns the Kubernetes client (exported method)
func (c *Collector) GetK8sClient() *k8s.Client {
	return c.k8s
}

// GetActiveAlertsCount returns the count of active alerts for an app
func (c *Collector) GetActiveAlertsCount(appID int64) (int, error) {
	query := `SELECT COUNT(*) FROM alerts WHERE app_id = ? AND status = 'active'`
	var count int
	err := c.db.GetConnection().QueryRow(query, appID).Scan(&count)
	return count, err
}

// GetTotalActiveAlerts returns the total count of active alerts across all apps
func (c *Collector) GetTotalActiveAlerts() (int, error) {
	query := `SELECT COUNT(*) FROM alerts WHERE status = 'active'`
	var count int
	err := c.db.GetConnection().QueryRow(query).Scan(&count)
	return count, err
}

func (c *Collector) getAppsToMonitor(appName string) ([]App, error) {
	var query string
	var args []interface{}

	if appName != "" {
		query = "SELECT id, name FROM apps WHERE name = ?"
		args = []interface{}{appName}
	} else {
		query = "SELECT id, name FROM apps"
	}

	rows, err := c.db.GetConnection().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var app App
		if err := rows.Scan(&app.ID, &app.Name); err != nil {
			continue
		}
		apps = append(apps, app)
	}

	return apps, nil
}