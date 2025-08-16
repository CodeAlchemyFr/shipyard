package monitoring

import (
	"time"
)

// MetricType represents the type of metric being collected
type MetricType string

const (
	MetricTypeCPU       MetricType = "cpu"
	MetricTypeMemory    MetricType = "memory"
	MetricTypeNetwork   MetricType = "network"
	MetricTypeDisk      MetricType = "disk"
	MetricTypePods      MetricType = "pods"
	MetricTypeRequests  MetricType = "requests"
	MetricTypeErrors    MetricType = "errors"
	MetricTypeLatency   MetricType = "latency"
)

// Metric represents a single metric data point
type Metric struct {
	ID        int64      `json:"id" db:"id"`
	AppID     int64      `json:"app_id" db:"app_id"`
	Type      MetricType `json:"metric_type" db:"metric_type"`
	Value     float64    `json:"value" db:"value"`
	Unit      string     `json:"unit" db:"unit"`
	PodName   string     `json:"pod_name" db:"pod_name"`
	Timestamp time.Time  `json:"timestamp" db:"timestamp"`
}

// HealthStatus represents the status of a health check
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusTimeout   HealthStatus = "timeout"
	HealthStatusError     HealthStatus = "error"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	ID           int64        `json:"id" db:"id"`
	AppID        int64        `json:"app_id" db:"app_id"`
	Endpoint     string       `json:"endpoint" db:"endpoint"`
	Method       string       `json:"method" db:"method"`
	Status       HealthStatus `json:"status" db:"status"`
	StatusCode   int          `json:"status_code" db:"status_code"`
	ResponseTime int          `json:"response_time" db:"response_time"` // milliseconds
	ErrorMessage string       `json:"error_message" db:"error_message"`
	CheckedAt    time.Time    `json:"checked_at" db:"checked_at"`
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// Alert represents a monitoring alert
type Alert struct {
	ID              int64         `json:"id" db:"id"`
	AppID           int64         `json:"app_id" db:"app_id"`
	Type            string        `json:"alert_type" db:"alert_type"`
	Threshold       float64       `json:"threshold" db:"threshold"`
	CurrentValue    float64       `json:"current_value" db:"current_value"`
	Severity        AlertSeverity `json:"severity" db:"severity"`
	Status          AlertStatus   `json:"status" db:"status"`
	Message         string        `json:"message" db:"message"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	ResolvedAt      *time.Time    `json:"resolved_at" db:"resolved_at"`
	AcknowledgedAt  *time.Time    `json:"acknowledged_at" db:"acknowledged_at"`
}

// Event represents a Kubernetes event
type Event struct {
	ID             int64     `json:"id" db:"id"`
	AppID          *int64    `json:"app_id" db:"app_id"` // nullable for cluster events
	Type           string    `json:"event_type" db:"event_type"`
	Reason         string    `json:"reason" db:"reason"`
	Message        string    `json:"message" db:"message"`
	ObjectKind     string    `json:"object_kind" db:"object_kind"`
	ObjectName     string    `json:"object_name" db:"object_name"`
	FirstTimestamp time.Time `json:"first_timestamp" db:"first_timestamp"`
	LastTimestamp  time.Time `json:"last_timestamp" db:"last_timestamp"`
	Count          int       `json:"count" db:"count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// MonitoringConfig represents monitoring configuration for an app
type MonitoringConfig struct {
	ID                     int64     `json:"id" db:"id"`
	AppID                  int64     `json:"app_id" db:"app_id"`
	Enabled                bool      `json:"enabled" db:"enabled"`
	HealthCheckPath        string    `json:"health_check_path" db:"health_check_path"`
	HealthCheckInterval    int       `json:"health_check_interval" db:"health_check_interval"` // seconds
	HealthCheckTimeout     int       `json:"health_check_timeout" db:"health_check_timeout"`   // seconds
	MetricsEnabled         bool      `json:"metrics_enabled" db:"metrics_enabled"`
	MetricsPath            string    `json:"metrics_path" db:"metrics_path"`
	MetricsPort            int       `json:"metrics_port" db:"metrics_port"`
	RetentionDays          int       `json:"retention_days" db:"retention_days"`
	CPUThreshold           float64   `json:"cpu_threshold" db:"cpu_threshold"`
	MemoryThreshold        float64   `json:"memory_threshold" db:"memory_threshold"`
	ErrorRateThreshold     float64   `json:"error_rate_threshold" db:"error_rate_threshold"`
	ResponseTimeThreshold  int       `json:"response_time_threshold" db:"response_time_threshold"` // milliseconds
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// AppMetrics represents aggregated metrics for an application
type AppMetrics struct {
	AppName        string    `json:"app_name"`
	CPUPercent     float64   `json:"cpu_percent"`
	MemoryPercent  float64   `json:"memory_percent"`
	MemoryBytes    int64     `json:"memory_bytes"`
	NetworkInBytes int64     `json:"network_in_bytes"`
	NetworkOutBytes int64    `json:"network_out_bytes"`
	PodCount       int       `json:"pod_count"`
	PodReady       int       `json:"pod_ready"`
	RequestsPerSec float64   `json:"requests_per_sec"`
	ErrorRate      float64   `json:"error_rate"`
	AvgLatency     float64   `json:"avg_latency"`
	LastUpdated    time.Time `json:"last_updated"`
}

// AppStatus represents the overall status of an application
type AppStatus struct {
	AppName       string        `json:"app_name"`
	Status        string        `json:"status"`        // healthy, warning, critical, unknown
	StatusIcon    string        `json:"status_icon"`   // 🟢, 🟡, 🔴, ⚪
	Replicas      string        `json:"replicas"`      // "3/5"
	LastDeploy    string        `json:"last_deploy"`   // "2h ago"
	ActiveAlerts  int           `json:"active_alerts"`
	Metrics       *AppMetrics   `json:"metrics,omitempty"`
	HealthCheck   *HealthCheck  `json:"health_check,omitempty"`
	RecentEvents  []Event       `json:"recent_events,omitempty"`
}

// ClusterStatus represents overall cluster health
type ClusterStatus struct {
	Healthy     bool   `json:"healthy"`
	NodesTotal  int    `json:"nodes_total"`
	NodesReady  int    `json:"nodes_ready"`
	PodsTotal   int    `json:"pods_total"`
	AlertsTotal int    `json:"alerts_total"`
	Message     string `json:"message"`
}